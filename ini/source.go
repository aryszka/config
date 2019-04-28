package ini

import "io"

type Source struct {
	name string
	root *node
}

func get(n *node, key []string) []string {
	if len(key) == 0 {
		n.used = true
		return n.values
	}

	child, ok := n.children[key[0]]
	if !ok {
		return nil
	}

	return get(child, key[1:])
}

func unused(n *node, key []string) [][]string {
	var u [][]string
	if len(n.values) > 0 && !n.used {
		u = append(u, key)
	}

	for symbol, child := range n.children {
		u = append(u, unused(child, append(key, symbol))...)
	}

	return u
}

func New(name string, input io.Reader) (*Source, error) {
	n, err := parse(input)
	if err != nil {
		return nil, err
	}

	root, err := postprocess(n)
	if err != nil {
		return nil, err
	}

	return &Source{name: name, root: root}, nil
}

func (s *Source) Name() string { return s.name }

func toInterfaces(s []string) []interface{} {
	var ifaces []interface{}
	for i := range s {
		ifaces = append(ifaces, s[i])
	}

	return ifaces
}

func (s *Source) Get(key ...string) []interface{} {
	values := get(s.root, key)
	return toInterfaces(values)
}

func (s *Source) Unused() [][]string {
	return unused(s.root, nil)
}
