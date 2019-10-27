package configsource

/*
// try writing examples for this first

type field struct {
}

// see: https://godoc.org/reflect#example-TypeOf
var configTargetType = reflect.TypeOf((*ConfigTarget)(nil)).Elem()

func fieldOf(t reflect.Type) (f field) {
	if t.Implements(configTargetType) {
		f.wildcard = true
		return
	}

	switch t.Kind() {
	case reflect.Slice:
		f.item = fieldOf(t.Elem())
	case reflect.Map:
		if keys are strings then check elems
		do we distinguish between map and slice items?
	case reflect.Struct:
		// all fields
		for fi in fields {
			f.fields = append(f.fields, fieldOf(fi))
		}
	case reflect.Interface:
		f.wildcard = true
	}

	return
}

func fieldsOf(target interface{}) []field {
}
*/
