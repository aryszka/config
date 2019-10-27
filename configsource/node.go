package configsource

/*
type node struct {
	flags  SourceFlags
	values []interface{}
	fields map[string]Source
}

var empty node

func (n node) Flags() SourceFlags {
	return n.flags
}

func (n node) Len() int {
	return len(n.values)
}

func (n node) Value(i int) interface{} {
	return n.values[i]
}

func (n node) Fields() []string {
	var f []string
	for fi := range n.fields {
		f = append(f, fi)
	}

	return f
}

func (n node) HasField(name string) bool {
	_, has := n.fields[name]
	return has
}

func (n node) Field(f string) Source {
	return n.fields[f]
}
*/
