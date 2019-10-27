package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type jsonLoader struct {
	input       io.Reader
	typeMapping map[NodeType]NodeType
}

func newJSONLoader(r io.Reader) jsonLoader {
	return jsonLoader{
		input: r,
		typeMapping: map[NodeType]NodeType{
			Float: Number,
		},
	}
}

func (l jsonLoader) Load() (interface{}, error) {
	b, err := ioutil.ReadAll(l.input)
	if err != nil {
		return nil, err
	}

	var o interface{}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (l jsonLoader) TypeMapping() map[NodeType]NodeType {
	return l.typeMapping
}

func JSON(r io.Reader) Source { return WithLoader(newJSONLoader(r)) }
