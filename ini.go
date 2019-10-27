package config

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
	if n.typ != Nil {
		return n.typ
	}

	return Primitive | List | Structure
}

func (n iniNode) Keys() []string {
	var keys []string
	for key := range n.ini.Fields {
		keys = append(keys, key)
	}

	return keys
}

func (s iniSource) Load() (Node, error) {
	n, err := ini.Read(s.input)
	if err != nil {
		return nil, err
	}

	return iniNode{ini: n}, nil
}

func INI(r io.Reader) Source { return iniSource{input: r} }
