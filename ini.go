package config

// It is possible that there are two targets, where one accepts a primitive value at a path, while the other
// accepts values from further down the path. This cannot be represented as a single object in Go, but it is
// possible to represent it in a single ini file. This is true for the flags, too. In fact, the behavior can be
// used for the positional arguments in case of the flags. Just as for subcommands, maybe.

// TODO: ensure that lists have no random order

import (
	"errors"
	"io"

	"github.com/aryszka/config/ini"
)

type iniNode struct {
	ini *ini.Node
	typ NodeType
}

type iniSource struct {
	input io.Reader
	done bool
	result Node
	err error
}

var errValuesAndFields = errors.New("values for a key with child keys not accepted")

func (n iniNode) Primitive() interface{} { return n.ini.Values[0] }
func (n iniNode) Len() int               { return len(n.ini.Values) }
func (n iniNode) Field(key string) Node  { return iniNode{ini: n.ini.Fields[key]} }

func (n iniNode) Item(i int) Node {
	return iniNode{
		ini: &ini.Node{Values: n.ini.Values[i : i+1]},
		typ: Primitive,
	}
}

func (n iniNode) Type() NodeType {
	if n.typ != undefined {
		return n.typ
	}

	return any
}

func (n iniNode) Keys() []string {
	var keys []string
	for key := range n.ini.Fields {
		keys = append(keys, key)
	}

	return keys
}

func (s *iniSource) Read() (Node, error) {
	if s.done {
		return s.result, s.err
	}

	s.done = true
	n, err := ini.Read(s.input)
	if err != nil {
		s.err = err
		return nil, err
	}

	s.result = iniNode{ini: n}
	return s.result, nil
}

func INI(r io.Reader) Source { return &iniSource{input: r} }
