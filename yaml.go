package config

import (
	"errors"
	"io"
	"io/ioutil"
	"unicode"

	"github.com/go-yaml/yaml"
)

type yamlLoader struct {
	input io.Reader
}

func invalidYAMLKey(...interface{}) error {
	return errors.New("invalid YAML key")
}

func newYAMLLoader(r io.Reader) Loader {
	return yamlLoader{input: r}
}

func sanitizeYAML(o interface{}) (interface{}, error) {
	switch ot := o.(type) {
	case []interface{}:
		for i := range ot {
			oi, err := sanitizeYAML(ot)
			if err != nil {
				return nil, err
			}

			ot[i] = oi
		}

		return o, nil
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key := range ot {
			skey, ok := key.(string)
			if !ok {
				return nil, invalidYAMLKey(key)
			}

			v, err := sanitizeYAML(ot[key])
			if err != nil {
				return nil, err
			}

			m[skey] = v
		}

		return m, nil
	default:
		return o, nil
	}
}

func (l yamlLoader) Load() (interface{}, error) {
	b, err := ioutil.ReadAll(l.input)
	if err != nil {
		return nil, err
	}

	r := []rune(string(b))
	ws := true
	for i := range r {
		if !unicode.IsSpace(r[i]) {
			ws = false
			break
		}
	}

	if ws {
		return nil, ErrEmptyConfig
	}

	var o interface{}
	if err := yaml.Unmarshal(b, &o); err != nil {
		return nil, err
	}

	return sanitizeYAML(o)
}

func (l yamlLoader) TypeMapping() map[NodeType]NodeType {
	return nil
}

func YAML(r io.Reader) Source { return WithLoader(newYAMLLoader(r)) }
