package config

import "errors"

type mergedNode struct {
	value      Node
	structures []Node
}

type mergedSource struct {
	sources []Source
}

type overrideSource struct {
	sources []Source
}

// TODO: optimize by memoizing, everywhere

func Merge(s ...Source) Source { return &mergedSource{sources: s} }

func (s mergedSource) Read() (Node, error) {
	var n []Node
	for _, si := range s.sources {
		ni, err := si.Read()
		if ni == nil && err == nil || errors.Is(err, ErrNoConfig) {
			continue
		}

		if err != nil {
			return nil, err
		}

		n = append(n, ni)
	}

	return mergeNodes(n...), nil
}

func mergeNodes(n ...Node) *mergedNode {
	// this merging can become interesting with map targets. What's the most expected?
	// the answer should go into a decision log and documentation

	var (
		valueNode  Node
		structures []Node
	)

	for _, ni := range n {
		t := ni.Type()
		if t&(Primitive|List) != 0 {
			valueNode = ni
		}

		if t&Structure != 0 {
			structures = append(structures, ni)
		}
	}

	return &mergedNode{value: valueNode, structures: structures}
}

func (n *mergedNode) Type() NodeType {
	t := Structure
	if n.value != nil {
		t |= n.value.Type()
	}

	return t
}

func (n *mergedNode) Primitive() interface{} { return n.value.Primitive() }
func (n *mergedNode) Len() int               { return n.value.Len() }
func (n *mergedNode) Item(i int) Node        { return n.value.Item(i) }

func (n *mergedNode) Keys() []string {
	var keys []string
	found := make(map[string]bool)
	for _, ni := range n.structures {
		for _, key := range ni.Keys() {
			if found[key] {
				continue
			}

			found[key] = true
			keys = append(keys, key)
		}
	}

	return keys
}

func (n *mergedNode) Field(key string) Node {
	var match []Node
	for _, ni := range n.structures {
		for _, ki := range ni.Keys() {
			if ki == key {
				match = append(match, ni.Field(key))
			}
		}
	}

	return mergeNodes(match...)
}

func Override(s ...Source) Source { return &overrideSource{sources: s} }

func (s overrideSource) Read() (Node, error) {
	for i := len(s.sources) - 1; i >= 0; i-- {
		n, err := s.sources[i].Read()
		if n == nil && err == nil || errors.Is(err, ErrNoConfig) {
			continue
		}

		return n, err
	}

	return nil, ErrNoConfig
}
