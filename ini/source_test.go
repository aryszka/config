package ini

import (
	"bytes"
	"testing"
)

func newSource(config string) (*Source, error) {
	return New("test.ini", bytes.NewBufferString(config))
}

func mustCreate(t *testing.T, config string) *Source {
	s, err := newSource(config)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func foo(t *testing.T) *Source {
	return mustCreate(t, "foo = 42")
}

func checkValue(t *testing.T, v []interface{}, expect ...interface{}) {
	if len(v) != len(expect) {
		t.Fatal("invalid length received", len(v), len(expect))
	}

	for i := range v {
		if v[i] != expect[i] {
			t.Fatal("invalid value received", i, v[i], expect[i])
		}
	}
}

func checkKeys(t *testing.T, k [][]string, expect ...[]string) {
	if len(k) != len(expect) {
		t.Fatal("invalid length received", len(k), len(expect))
	}

	for i := range k {
		if len(k[i]) != len(expect[i]) {
			t.Fatal("invalid key length received", i, len(k[i]), len(expect[i]))
		}

		for j := range k[i] {
			if k[i][j] != expect[i][j] {
				t.Fatal("invalid key received", i, j, k[i][j], expect[i][j])
			}
		}
	}
}

func TestName(t *testing.T) {
	s, err := newSource("")
	if err != nil {
		t.Fatal(err)
	}

	name := s.Name()
	if name != "test.ini" {
		t.Error("invalid name", name)
	}
}

func TestRead(t *testing.T) {
	t.Run("parse failed", func(t *testing.T) {
		if _, err := newSource("invalid"); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("parse successful", func(t *testing.T) {
		if _, err := newSource("foo = 42"); err != nil {
			t.Error(err)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s := foo(t)
		v := s.Get("bar")
		checkValue(t, v)
	})

	t.Run("get simple", func(t *testing.T) {
		s := foo(t)
		v := s.Get("foo")
		checkValue(t, v, "42")
	})

	t.Run("composite", func(t *testing.T) {
		t.Run("get", func(t *testing.T) {
			s := mustCreate(t, "foo.bar.baz = 42")
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42")
		})

		t.Run("short", func(t *testing.T) {
			s := mustCreate(t, "foo.bar.baz = 42")
			v := s.Get("foo", "bar")
			checkValue(t, v)
		})

		t.Run("long", func(t *testing.T) {
			s := mustCreate(t, "foo.bar.baz = 42")
			v := s.Get("foo", "bar", "baz", "qux")
			checkValue(t, v)
		})
	})

	t.Run("multiple values", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			s := mustCreate(t, `
				foo = 42
				foo = 84
				bar = 120
			`)
			v := s.Get("foo")
			checkValue(t, v, "42", "84")
		})

		t.Run("composite", func(t *testing.T) {
			s := mustCreate(t, `
				foo.bar.baz = 42
				foo.bar.baz = 84
				foo.bar = 120
				foo.bar.baz.qux = 240
			`)
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42", "84")
		})

		t.Run("spread", func(t *testing.T) {
			s := mustCreate(t, `
				foo.bar.baz = 42
				foo.bar = 120
				foo.bar.baz = 84
				foo.bar.baz.qux = 240
			`)
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42", "84")
		})
	})

	t.Run("group", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			s := mustCreate(t, `
				[foo]
				42
			`)
			v := s.Get("foo")
			checkValue(t, v, "42")
		})

		t.Run("composite", func(t *testing.T) {
			s := mustCreate(t, `
				[foo.bar]
				baz = 42
			`)
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42")
		})

		t.Run("short", func(t *testing.T) {
			s := mustCreate(t, `
				[foo.bar]
				baz = 42
			`)
			v := s.Get("foo", "bar")
			checkValue(t, v)
		})

		t.Run("long", func(t *testing.T) {
			s := mustCreate(t, `
				[foo.bar]
				baz = 42
			`)
			v := s.Get("foo", "bar", "baz", "qux")
			checkValue(t, v)
		})

		t.Run("multiple values", func(t *testing.T) {
			s := mustCreate(t, `
				[foo.bar]
				baz = 42
				baz = 84
			`)
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42", "84")
		})

		t.Run("multiple values, spread", func(t *testing.T) {
			s := mustCreate(t, `
				[foo.bar]
				baz = 42

				foo.bar = 120

				foo.bar.baz = 84

				[foo.bar.baz]
				240

				[foo]
				360
			`)
			v := s.Get("foo", "bar", "baz")
			checkValue(t, v, "42", "84", "240")
		})
	})

	t.Run("quoting", func(t *testing.T) {
		t.Run("single", func(t *testing.T) {
			s := mustCreate(t, "foo = 'bar'")
			v := s.Get("foo")
			checkValue(t, v, "bar")
		})

		t.Run("double", func(t *testing.T) {
			s := mustCreate(t, "foo = \"bar\"")
			v := s.Get("foo")
			checkValue(t, v, "bar")
		})
	})

	t.Run("escaping", func(t *testing.T) {
		t.Run("not quoted", func(t *testing.T) {
			s := mustCreate(t, "foo = \\\n\\'\\\"\\\\\\[\\]\\=\\#")
			v := s.Get("foo")
			checkValue(t, v, "\n'\"\\[]=#")
		})

		t.Run("single quoted", func(t *testing.T) {
			s := mustCreate(t, "foo = '\\\\\\''")
			v := s.Get("foo")
			checkValue(t, v, "\\'")
		})

		t.Run("double quoted", func(t *testing.T) {
			s := mustCreate(t, "foo = \"\\\\\\\"\"")
			v := s.Get("foo")
			checkValue(t, v, "\\\"")
		})
	})
}

func TestUnused(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		s := mustCreate(t, `
			foo.bar.baz = 42
			foo.bar.qux = 84
		`)
		s.Get("foo", "bar", "baz")
		s.Get("foo", "bar", "qux")
		k := s.Unused()
		checkKeys(t, k)
	})

	t.Run("part", func(t *testing.T) {
		s := mustCreate(t, `
			foo.bar.baz = 42
			foo.bar.qux = 84
		`)
		s.Get("foo", "bar", "baz")
		k := s.Unused()
		checkKeys(t, k, []string{"foo", "bar", "qux"})
	})

	t.Run("all", func(t *testing.T) {
		s := mustCreate(t, `
			foo.bar.baz = 42
			foo.bar.qux = 84
		`)
		k := s.Unused()
		checkKeys(t, k, []string{"foo", "bar", "baz"}, []string{"foo", "bar", "qux"})
	})

	t.Run("empty", func(t *testing.T) {
		s := mustCreate(t, "")
		k := s.Unused()
		checkKeys(t, k)
	})
}
