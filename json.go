package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type jsonReader struct {
	input       io.Reader
	typeMapping map[NodeType]NodeType
}

func newJSONReader(r io.Reader) *jsonReader {
	return &jsonReader{
		input: r,
		typeMapping: map[NodeType]NodeType{
			Float: Number,
		},
	}
}

func (l *jsonReader) Read() (interface{}, error) {
	b, err := ioutil.ReadAll(l.input)
	if err != nil {
		return nil, err
	}

	var o interface{}
	err = json.Unmarshal(b, &o)
	return o, err
}

func (l jsonReader) TypeMapping() map[NodeType]NodeType {
	return l.typeMapping
}

func JSON(r io.Reader) Source { return WithReader(newJSONReader(r)) }
