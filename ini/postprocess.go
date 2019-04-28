package ini

import "fmt"

type node struct {
	values   []string
	children map[string]*node
	used     bool
}

func errUnexpectedASTNodeType(name string) error {
	return fmt.Errorf("unexpected AST node type: %s", name)
}

func errUnexpectedASTNodeStructure(n *Node) error {
	return fmt.Errorf("unexpected AST node structure: %s", n.Name)
}

func processQuote(parent *node, n *Node) error {
	text, err := unquote(n.Text())
	if err != nil {
		return err
	}

	parent.values = append(parent.values, text)
	return nil
}

func processValue(parent *node, n *Node) error {
	if len(n.Nodes) == 1 {
		return processNode(parent, n.Nodes[0])
	}

	text, err := unescapeNonQuote(n.Text())
	if err != nil {
		return err
	}

	parent.values = append(parent.values, text)
	return nil
}

func getKey(n *Node) []string {
	var key []string
	for _, symbol := range n.Nodes {
		key = append(key, symbol.Text())
	}

	return key
}

func getOrCreateChild(n *node, key []string) *node {
	if len(key) == 0 {
		return n
	}

	child, exists := n.children[key[0]]
	if !exists {
		child = &node{children: make(map[string]*node)}
		n.children[key[0]] = child
	}

	return getOrCreateChild(child, key[1:])
}

func processKeyedValue(parent *node, n *Node) error {
	if len(n.Nodes) < 2 {
		return errUnexpectedASTNodeStructure(n)
	}

	key := getKey(n.Nodes[0])
	child := getOrCreateChild(parent, key)
	return processNode(child, n.Nodes[1])
}

func processNodes(parent *node, n []*Node) error {
	for i := range n {
		if err := processNode(parent, n[i]); err != nil {
			return err
		}
	}

	return nil
}

func processGroup(parent *node, n *Node) error {
	if len(n.Nodes) == 0 || len(n.Nodes[0].Nodes) == 0 {
		return errUnexpectedASTNodeStructure(n)
	}

	key := getKey(n.Nodes[0].Nodes[0])
	child := getOrCreateChild(parent, key)
	return processNodes(child, n.Nodes[1:])
}

func processConfig(parent *node, n *Node) error {
	return processNodes(parent, n.Nodes)
}

func processNode(parent *node, n *Node) error {
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
		return errUnexpectedASTNodeType(n.Name)
	}
}

func postprocess(n *Node) (*node, error) {
	root := &node{children: make(map[string]*node)}
	err := processNode(root, n)
	return root, err
}
