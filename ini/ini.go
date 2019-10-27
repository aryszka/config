package ini

import (
	"io"

	"github.com/aryszka/config/ini/syntax"
)

type Node struct {
	Values []string
	Fields map[string]*Node
}

func Read(r io.Reader) (*Node, error) {
	ast, err := syntax.Parse(r)
	if err != nil {
		return nil, err
	}

	return postprocess(ast)
}
