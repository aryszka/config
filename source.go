package config

import (
	"errors"
	"fmt"
)

// TODO: gradual loader may be required because when loading from a database, we may not want to load everything
// if it's not required. An additional loader node type may be the right solution. This will be more interesting
// when it will be possible to update the config on-the-fly.

type Loader interface {
	Load() (interface{}, error)
	TypeMapping() map[NodeType]NodeType
}

type NodeType int

const (
	Nil  NodeType = 0
	Bool NodeType = 1 << iota
	Int
	Float
	String
	List
	Structure
	ignored   // for testing
	Number    = Int | Float
	Primitive = Bool | Number | String
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
	Load() (Node, error)
}

// TODO: split source and node
type source struct {
	loader      Loader
	typeMapping map[NodeType]NodeType
	node        interface{}
	name        string
}

var (
	ErrLoaderImplementation = errors.New("loader implementation")
	ErrEmptyConfig          = errors.New("empty config")
	ErrInvalidTarget        = errors.New("invalid target")
	ErrInvalidInputValue    = errors.New("invalid input value")
	ErrTooManyValues        = errors.New("too many values")
	ErrNumericOverflow      = fmt.Errorf("%w: integer overflow", ErrInvalidInputValue)
	ErrConflictingKeys      = errors.New("conflicting keys")
)

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

func WithLoader(l Loader) Source {
	return source{loader: l}
}

func (s source) Load() (Node, error) {
	node, err := s.loader.Load()
	if err != nil {
		return nil, s.sourceError(err)
	}

	return source{node: node, typeMapping: s.loader.TypeMapping()}, nil
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
			ErrLoaderImplementation,
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
