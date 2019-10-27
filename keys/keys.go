package keys

import "github.com/iancoleman/strcase"

func CanonicalSymbol(symbol string) string {
	return strcase.ToKebab(symbol)
}

func Canonical(key ...string) []string {
	var can []string
	for i := range key {
		can = append(can, CanonicalSymbol(key[i]))
	}

	return can
}
