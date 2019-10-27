package configsource

/*
import "encoding/json"

type jsonFile struct {
	file
}

func JSON(name string) Loader {
	return jsonFile{file: file{name: name}}
}

func jsonStruct(s map[string]interface{}) Source {
	var n node
	if len(s) == 0 {
		return n
	}

	n.fields = make(map[string]Source)
	for k, v := range s {
		n.fields[k] = jsonToSource(v)
	}

	return n
}

func jsonList(l []interface{}) Source {
	var n node
	for _, v := range l {
		n.values = append(n.values, jsonToSource(v))
	}

	return n
}

func jsonValue(v interface{}) Source {
	return node{values: []interface{}{v}}
}

func jsonToSource(doc interface{}) Source {
	switch v := doc.(type) {
	case map[string]interface{}:
		return jsonStruct(v)
	case []interface{}:
		return jsonList(v)
	default:
		return jsonValue(v)
	}
}

func (j jsonFile) Load() (Source, error) {
	b, err := loadFile(j.file.name)
	if err != nil {
		return nil, err
	}

	var doc interface{}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}

	return jsonToSource(doc), nil
}
*/
