package ini

import (
	"errors"

	"github.com/aryszka/config/ini/syntax"
)

var errUnexpectedParserResult = errors.New("unexpected parser result")

func processQuote(parent *Node, n *syntax.Node) error {
	text, err := unquote(n.Text())
	if err != nil {
		return err
	}

	parent.Values = append(parent.Values, text)
	return nil
}

func processValue(parent *Node, n *syntax.Node) error {
	if len(n.Nodes) > 0 {
		return processNode(parent, n.Nodes[0])
	}

	text, err := unescapeNonQuote(n.Text())
	if err != nil {
		return err
	}

	parent.Values = append(parent.Values, text)
	return nil
}

func getKey(n *syntax.Node) []string {
	var key []string
	for _, symbol := range n.Nodes {
		key = append(key, symbol.Text())
	}

	return key
}

func getOrCreateChild(n *Node, key []string) *Node {
	if len(key) == 0 {
		return n
	}

	if n.Fields == nil {
		n.Fields = make(map[string]*Node)
	}

	child, exists := n.Fields[key[0]]
	if !exists {
		child = &Node{}
		n.Fields[key[0]] = child
	}

	return getOrCreateChild(child, key[1:])
}

func processKeyedValue(parent *Node, n *syntax.Node) error {
	if len(n.Nodes) < 2 {
		// TODO: error info
		return errUnexpectedParserResult
	}

	key := getKey(n.Nodes[0])
	child := getOrCreateChild(parent, key)
	return processNode(child, n.Nodes[1])
}

func processNodes(parent *Node, n []*syntax.Node) error {
	for i := range n {
		if err := processNode(parent, n[i]); err != nil {
			return err
		}
	}

	return nil
}

func processGroup(parent *Node, n *syntax.Node) error {
	if len(n.Nodes) == 0 || len(n.Nodes[0].Nodes) == 0 {
		// TODO: error info
		return errUnexpectedParserResult
	}

	key := getKey(n.Nodes[0].Nodes[0])
	child := getOrCreateChild(parent, key)
	return processNodes(child, n.Nodes[1:])
}

func processConfig(parent *Node, n *syntax.Node) error {
	return processNodes(parent, n.Nodes)
}

func processNode(parent *Node, n *syntax.Node) error {
	switch n.Name {
	case "quote":
		return processQuote(parent, n)
	case "value":
		return processValue(parent, n)
	case "keyed-value":
		return processKeyedValue(parent, n)
	case "group":
		return processGroup(parent, n)
	case "config":
		return processConfig(parent, n)
	default:
		// TODO: error info
		return errUnexpectedParserResult
	}
}

func postprocess(n *syntax.Node) (*Node, error) {
	root := &Node{}
	err := processNode(root, n)
	return root, err
}
