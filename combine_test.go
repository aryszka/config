package config

import (
	"bytes"
	"errors"
	"testing"
)

type testSource struct {
	node Node
	err  error
}

func (s testSource) Read() (Node, error) {
	return s.node, s.err
}

func jsonString(j string) Source {
	return JSON(bytes.NewBufferString(j))
}

func iniString(i string) Source {
	return INI(bytes.NewBufferString(i))
}

func TestMerge(t *testing.T) {
	t.Run("invalid no node", func(t *testing.T) {
		var o struct{ Foo int }
		s := Merge(
			testSource{},
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to merge sources")
		}
	})

	t.Run("no config", func(t *testing.T) {
		var o struct{ Foo int }
		s := Merge(
			testSource{err: ErrNoConfig},
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to merge sources")
		}
	})

	t.Run("error", func(t *testing.T) {
		var o struct{ Foo int }
		s := Merge(
			testSource{err: errors.New("test error")},
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("no config at all", func(t *testing.T) {
		t.Run("structure", func(t *testing.T) {
			o := struct{ Foo int }{42}
			s := Merge(
				testSource{err: ErrNoConfig},
				testSource{err: ErrNoConfig},
			)

			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to ignore sources")
			}
		})

		t.Run("primitive", func(t *testing.T) {
			o := 42
			s := Merge(
				testSource{err: ErrNoConfig},
				testSource{err: ErrNoConfig},
			)

			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to ignore sources")
			}
		})
	})

	t.Run("only primitive", func(t *testing.T) {
		var o int
		s := Merge(
			jsonString("21"),
			jsonString("42"),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != 42 {
			t.Error("failed to merge sources")
		}
	})

	t.Run("only structures", func(t *testing.T) {
		var o struct{ Foo int }
		s := Merge(
			jsonString(`{"foo": 21}`),
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to merge sources")
		}
	})

	t.Run("primitive and structure", func(t *testing.T) {
		var (
			o1 struct{Foo int}
			o2 struct{Foo struct{Bar int}}
		)

		s := Merge(
			iniString("foo.bar=21"),
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o1, s); err != nil {
			t.Error(err)
		}

		if o1.Foo != 42 {
			t.Error("failed to merge sources")
		}

		if err := Apply(&o2, s); err != nil {
			t.Error(err)
		}

		if o2.Foo.Bar != 21 {
			t.Error("failed to merge sources")
		}
	})
}

func TestOverride(t *testing.T) {
	t.Run("invalid no node", func(t *testing.T) {
		var o struct{ Foo int }
		s := Override(
			testSource{},
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to override sources")
		}
	})

	t.Run("no config", func(t *testing.T) {
		var o struct{ Foo int }
		s := Override(
			jsonString(`{"foo": 42}`),
			testSource{err: ErrNoConfig},
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to override sources")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Run("overridden", func(t *testing.T) {
			var o struct{ Foo int }
			s := Override(
				testSource{err: errors.New("test error")},
				jsonString(`{"foo": 42}`),
			)

			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to override sources")
			}
		})

		t.Run("evaluated", func(t *testing.T) {
			var o struct{ Foo int }
			s := Override(
				jsonString(`{"foo": 42}`),
				testSource{err: errors.New("test error")},
			)

			if err := Apply(&o, s); err == nil {
				t.Error("failed to fail")
			}
		})
	})

	t.Run("no config at all", func(t *testing.T) {
		o := struct{ Foo int }{42}
		s := Override(
			testSource{err: ErrNoConfig},
			testSource{err: ErrNoConfig},
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to ignore sources")
		}
	})

	t.Run("last used", func(t *testing.T) {
		var o struct{Foo int}
		s := Override(
			jsonString(`{"foo": 21}`),
			jsonString(`{"foo": 42}`),
		)

		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to merge sources")
		}
	})
}
