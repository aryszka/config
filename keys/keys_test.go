package keys

import "testing"

func TestCanonical(t *testing.T) {
	t.Run("symbol", func(t *testing.T) {
		c := CanonicalSymbol("FooBarBaz")
		if c != "foo-bar-baz" {
			t.Error("failed to canonicalize symbol", c)
		}
	})

	t.Run("key", func(t *testing.T) {
		c := Canonical("FooBarBaz", "QuxQuzQuuz")
		if len(c) != 2 || c[0] != "foo-bar-baz" || c[1] != "qux-quz-quuz" {
			t.Error("failed to canonicalize key", c)
		}
	})
}
