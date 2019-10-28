package config

import (
	"errors"
	"fmt"
)

// TODO: gradual reader may be required because when reading from a database, we may not want to read everything
// if it's not required. An additional reader node type may be the right solution. This will be more interesting
// when it will be possible to update the config on-the-fly.

type Reader interface {
	Read() (interface{}, error)
	TypeMapping() map[NodeType]NodeType
}

type NodeType int

const undefined NodeType = 0

const (
	Nil NodeType = 1 << iota
	Bool
	Int
	Float
	String
	List
	Structure
	ignored   // for testing
	Number    = Int | Float
	Primitive = Bool | Number | String
	any       = Primitive | List | Structure
)

type Node interface {
	Type() NodeType
	Primitive() interface{}
	Len() int
	Item(int) Node
	Keys() []string
	Field(string) Node
}

type Source interface {
	Read() (Node, error)
}

// TODO: split source and node
type source struct {
	reader      Reader
	typeMapping map[NodeType]NodeType
	node        interface{}
	name        string
	hasRead bool
	err error
}

var (
	ErrSourceImplementation = errors.New("reader implementation")
	ErrNoConfig             = errors.New("empty config")
	ErrInvalidTarget        = errors.New("invalid target")
	ErrInvalidInputValue    = errors.New("invalid input value")
	ErrTooManyValues        = errors.New("too many values")
	ErrNumericOverflow      = fmt.Errorf("%w: integer overflow", ErrInvalidInputValue)
	ErrConflictingKeys      = errors.New("conflicting keys")
)

func WithReader(l Reader) Source {
	return &source{reader: l}
}

func (s source) sourceErrorf(format string, args ...interface{}) error {
	if s.name != "" {
		format = "source=%s; " + format
		args = append([]interface{}{s.name}, args...)
	}

	return fmt.Errorf(
		format,
		args...,
	)
}

func (s source) sourceError(err error) error {
	return s.sourceErrorf("%w", err)
}

func (s *source) Read() (Node, error) {
	if s.hasRead && s.err != nil {
		return nil, s.err
	}

	if s.hasRead {
		return source{node: s.node, typeMapping: s.reader.TypeMapping()}, nil
	}

	s.hasRead = true
	node, err := s.reader.Read()
	if err != nil {
		s.err = s.sourceError(err)
		return nil, s.err
	}

	s.node = node
	return source{node: node, typeMapping: s.reader.TypeMapping()}, nil
}

func (s source) defaultType() NodeType {
	switch s.node.(type) {
	case nil:
		return Nil
	case bool:
		return Bool
	case int,
		int8,
		int16,
		int32,
		int64,
		uint,
		uint8,
		uint16,
		uint32,
		uint64:
		return Int
	case float64, float32:
		return Float
	case string:
		return String
	case []interface{}:
		return List
	case map[string]interface{}:
		return Structure
	default:
		panic(s.sourceErrorf(
			"%w; unexpected type: %v",
			ErrSourceImplementation,
			s.node,
		))
	}
}

func (s source) Type() NodeType {
	dt := s.defaultType()
	t, ok := s.typeMapping[dt]
	if !ok {
		return dt
	}

	return t
}

func (s source) Primitive() interface{} {
	return s.node
}

func (s source) Len() int {
	return len(s.node.([]interface{}))
}

func (s source) Item(i int) Node {
	return source{node: s.node.([]interface{})[i], typeMapping: s.typeMapping}
}

func (s source) Keys() []string {
	var keys []string
	for key := range s.node.(map[string]interface{}) {
		keys = append(keys, key)
	}

	return keys
}

func (s source) Field(key string) Node {
	return source{node: s.node.(map[string]interface{})[key], typeMapping: s.typeMapping}
}
