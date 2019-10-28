package config

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"
)

type testLoader struct {
	value interface{}
	fail  bool
	types map[NodeType]NodeType
}

var (
	errTestLoadFailed = errors.New("test load failed")
	errOddMapping     = errors.New("odd mapping")
)

func singleValueLoader(value interface{}, mapping ...NodeType) Loader {
	if len(mapping)%2 != 0 {
		panic(errOddMapping)
	}

	types := make(map[NodeType]NodeType)
	for i := 0; i < len(mapping); i += 2 {
		types[mapping[i]] = mapping[i+1]
	}

	return testLoader{value: value, types: types}
}

func failingLoader() Loader {
	return testLoader{fail: true}
}

func (l testLoader) Load() (interface{}, error) {
	if l.fail {
		return nil, errTestLoadFailed
	}

	return l.value, nil
}

func (l testLoader) TypeMapping() map[NodeType]NodeType {
	return l.types
}

func TestApplyInvalidTarget(t *testing.T) {
	t.Run("zero options", func(t *testing.T) {
		type options struct{ Foo int }
		var o options
		j := bytes.NewBufferString(`{"foo": 42}`)
		s := JSON(j)
		if err := Apply(o, s); !errors.Is(err, ErrInvalidTarget) {
			t.Error("failed to fail with the right error", err)
		}
	})

	t.Run("options with defaults", func(t *testing.T) {
		type options struct{ Foo int }
		o := options{Foo: 42}
		j := bytes.NewBufferString(`{"foo": 42}`)
		s := JSON(j)
		if err := Apply(o, s); !errors.Is(err, ErrInvalidTarget) {
			t.Error("failed to fail with the right error", err)
		}

		if o.Foo != 42 {
			t.Error("failed to leave defaults")
		}
	})
}

func TestApplyEmptyConfig(t *testing.T) {
	t.Run("empty options", func(t *testing.T) {
		type options struct{}
		var o options
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}
	})

	t.Run("options with fields", func(t *testing.T) {
		type options struct{ Foo int }
		var o options
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 0 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("options with fields, with defaults", func(t *testing.T) {
		type options struct{ Foo int }
		o := options{Foo: 42}
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("list options", func(t *testing.T) {
		var o []interface{}
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 0 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("list options, with default items", func(t *testing.T) {
		o := []int{42}
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 1 || o[0] != 42 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("map options", func(t *testing.T) {
		var o map[string]interface{}
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 0 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("map options, with defaults", func(t *testing.T) {
		o := map[string]int{"foo": 42}
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 1 || o["foo"] != 42 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("primitive option", func(t *testing.T) {
		var o int
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != 0 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("primitive option, with default value", func(t *testing.T) {
		o := 42
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != 42 {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("interface option", func(t *testing.T) {
		type iface interface{}
		var o iface
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != nil {
			t.Error("failed to leave defaults")
		}
	})

	t.Run("interface option, with defaults", func(t *testing.T) {
		type iface interface{}
		o := 42
		y := bytes.NewBuffer(nil)
		s := YAML(y)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != 42 {
			t.Error("failed to leave defaults")
		}
	})
}

func TestApplyLoadFailed(t *testing.T) {
	type options struct{ Foo int }
	o := options{Foo: 42}
	s := WithLoader(failingLoader())
	if err := Apply(&o, s); !errors.Is(err, errTestLoadFailed) {
		t.Error("failed to fail with the right error", err)
	}

	if o.Foo != 42 {
		t.Error("failed to leave defaults")
	}
}

func TestApplyToBool(t *testing.T) {
	t.Run("not bool", func(t *testing.T) {
		var o bool
		j := bytes.NewBufferString("42")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("parsed, lying loader", func(t *testing.T) {
		var o bool
		s := WithLoader(singleValueLoader(42, Int, Bool))
		if err := Apply(&o, s); !errors.Is(err, ErrLoaderImplementation) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("not parsed, not bool", func(t *testing.T) {
		var o struct{ Foo bool }
		i := bytes.NewBufferString("foo=42")
		s := INI(i)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error", err)
		}
	})

	t.Run("parsed", func(t *testing.T) {
		var o bool
		j := bytes.NewBufferString("true")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if !o {
			t.Error("failed to apply bool")
		}
	})

	t.Run("not parsed", func(t *testing.T) {
		var o struct{ Foo bool }
		i := bytes.NewBufferString("foo=true")
		s := INI(i)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if !o.Foo {
			t.Error("failed to apply bool")
		}
	})

	t.Run("optionally list, more than once", func(t *testing.T) {
		var o struct{ Foo bool }
		i := bytes.NewBufferString("foo=true\nfoo=false")
		s := INI(i)
		if err := Apply(&o, s); !errors.Is(err, ErrTooManyValues) {
			t.Error("failed to fail with the right error", err)
		}
	})

	t.Run("optionally list, no items", func(t *testing.T) {
		o := struct{ Foo bool }{true}
		i := bytes.NewBufferString("foo.bar=1\nfoo.baz=2")
		s := INI(i)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if !o.Foo {
			t.Error("failed to ignore optional list")
		}
	})

	t.Run("has default", func(t *testing.T) {
		o := struct{ Foo bool }{true}
		i := bytes.NewBufferString("foo=false")
		s := INI(i)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo {
			t.Error("failed to apply bool")
		}
	})
}

func TestApplyToInt(t *testing.T) {
	t.Run("not int", func(t *testing.T) {
		var o int
		j := bytes.NewBufferString("true")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error", err)
		}
	})

	t.Run("optionally list", func(t *testing.T) {
		t.Run("too many items", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=1\nfoo=2")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrTooManyValues) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("no items", func(t *testing.T) {
			o := struct{ Foo int }{42}
			i := bytes.NewBufferString("foo.bar=1\nfoo.baz=2")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to ignore optional list")
			}
		})

		t.Run("one item", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to ignore optional list")
			}
		})
	})

	t.Run("not parsed", func(t *testing.T) {
		t.Run("not int", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=true")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("sign only", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=-")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("hexa prefix only", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=0x")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("overflow max bits", func(t *testing.T) {
			var o struct{ Foo int64 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=%d", ^uint64(0)>>1+1))
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("overflow max bits, negative", func(t *testing.T) {
			var o struct{ Foo int64 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=-%d", ^uint64(0)>>1+2))
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("overflow, 16bit", func(t *testing.T) {
			var o struct{ Foo int16 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=%d", ^uint16(0)>>1+1))
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("overflow, 16bit, negative", func(t *testing.T) {
			var o struct{ Foo int16 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=-%d", ^uint16(0)>>1+2))
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("hexa zero", func(t *testing.T) {
			o := struct{ Foo int }{42}
			i := bytes.NewBufferString("foo=0x00")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("hexa", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=0x2a")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("hexa, negative", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=-0x2a")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != -42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("octal zero", func(t *testing.T) {
			o := struct{ Foo int }{42}
			i := bytes.NewBufferString("foo=000")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("octal", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=052")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("octal, negative", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=-052")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != -42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("zero", func(t *testing.T) {
			o := struct{ Foo int }{42}
			i := bytes.NewBufferString("foo=0")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("decimal", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("decimal, negative", func(t *testing.T) {
			var o struct{ Foo int }
			i := bytes.NewBufferString("foo=-42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != -42 {
				t.Error("failed to apply int value")
			}
		})

		t.Run("max bits", func(t *testing.T) {
			var o struct{ Foo int64 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=%d", ^uint64(0)>>1))
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != int64(^uint64(0)>>1) {
				t.Error("failed to apply int")
			}
		})

		t.Run("max bits, negative", func(t *testing.T) {
			var o struct{ Foo int64 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=-%d", ^uint64(0)>>1+1))
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0-int64(^uint64(0)>>1)-1 {
				t.Error("failed to apply int")
			}
		})

		t.Run("has default", func(t *testing.T) {
			o := struct{ Foo int }{21}
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply int")
			}
		})
	})

	t.Run("parsed", func(t *testing.T) {
		t.Run("overflow", func(t *testing.T) {
			var o int16
			s := WithLoader(singleValueLoader(int(^uint16(0)>>1) + 1))
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("native size", func(t *testing.T) {
			var o int
			y := bytes.NewBufferString("42")
			s := YAML(y)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply int")
			}
		})

		t.Run("fixed size", func(t *testing.T) {
			var o int16
			y := bytes.NewBufferString("42")
			s := YAML(y)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply int")
			}
		})

		t.Run("unsigned overflow", func(t *testing.T) {
			var o int16
			s := WithLoader(singleValueLoader(^uint16(0)>>1 + 1))
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("unsigned, no overflow", func(t *testing.T) {
			var o int16
			s := WithLoader(singleValueLoader(uint64(42)))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply int")
			}
		})

		t.Run("float, not round", func(t *testing.T) {
			var o int
			j := bytes.NewBufferString("3.14")
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("float, overflow", func(t *testing.T) {
			var o int16
			j := bytes.NewBufferString(fmt.Sprint(^uint16(0)>>1 + 1))
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("float", func(t *testing.T) {
			var o int
			j := bytes.NewBufferString("42")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply int")
			}
		})

		t.Run("lying loader", func(t *testing.T) {
			var o int
			s := WithLoader(singleValueLoader(true, Bool, Int))
			if err := Apply(&o, s); !errors.Is(err, ErrLoaderImplementation) {
				t.Error("failed to fail with the right error")
			}
		})
	})
}

func TestApplyToUint(t *testing.T) {
	t.Run("not uint", func(t *testing.T) {
		var o uint
		j := bytes.NewBufferString("true")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error", err)
		}
	})

	t.Run("optionally list", func(t *testing.T) {
		t.Run("too many items", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=1\nfoo=2")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrTooManyValues) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("no items", func(t *testing.T) {
			o := struct{ Foo uint }{42}
			i := bytes.NewBufferString("foo.bar=1\nfoo.baz=2")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to ignore optional list")
			}
		})

		t.Run("one item", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to ignore optional list")
			}
		})
	})

	t.Run("not parsed", func(t *testing.T) {
		t.Run("not int", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=true")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("hexa prefix only", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=0x")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("overflow, 16bit", func(t *testing.T) {
			var o struct{ Foo uint16 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=%d", int(^uint16(0))+1))
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("hexa zero", func(t *testing.T) {
			o := struct{ Foo uint }{42}
			i := bytes.NewBufferString("foo=0x00")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("hexa", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=0x2a")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("octal zero", func(t *testing.T) {
			o := struct{ Foo uint }{42}
			i := bytes.NewBufferString("foo=000")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("octal", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=052")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("zero", func(t *testing.T) {
			o := struct{ Foo uint }{42}
			i := bytes.NewBufferString("foo=0")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 0 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("decimal", func(t *testing.T) {
			var o struct{ Foo uint }
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply uint value")
			}
		})

		t.Run("max bits", func(t *testing.T) {
			var o struct{ Foo uint64 }
			i := bytes.NewBufferString(fmt.Sprintf("foo=%d", ^uint64(0)))
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != ^uint64(0) {
				t.Error("failed to apply uint")
			}
		})

		t.Run("has default", func(t *testing.T) {
			o := struct{ Foo uint }{21}
			i := bytes.NewBufferString("foo=42")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 42 {
				t.Error("failed to apply uint")
			}
		})
	})

	t.Run("parsed", func(t *testing.T) {
		t.Run("overflow", func(t *testing.T) {
			var o uint16
			s := WithLoader(singleValueLoader(uint(^uint16(0)) + 1))
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("native size", func(t *testing.T) {
			var o uint
			s := WithLoader(singleValueLoader(uint(42)))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply uint")
			}
		})

		t.Run("fixed size", func(t *testing.T) {
			var o uint16
			s := WithLoader(singleValueLoader(uint(42)))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply uint")
			}
		})

		t.Run("signed, overflow", func(t *testing.T) {
			var o uint16
			s := WithLoader(singleValueLoader(int(^uint16(0)) + 1))
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("signed, negative overflow", func(t *testing.T) {
			var o uint
			y := bytes.NewBufferString("-42")
			s := YAML(y)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("unsigned, no overflow", func(t *testing.T) {
			var o uint16
			s := WithLoader(singleValueLoader(int64(42)))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply uint")
			}
		})

		t.Run("float, not round", func(t *testing.T) {
			var o uint
			j := bytes.NewBufferString("3.14")
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("float, overflow", func(t *testing.T) {
			var o uint16
			j := bytes.NewBufferString(fmt.Sprint(int(^uint16(0)) + 1))
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("float, negative overflow", func(t *testing.T) {
			var o uint
			j := bytes.NewBufferString("-42")
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error", err)
			}
		})

		t.Run("float", func(t *testing.T) {
			var o uint
			j := bytes.NewBufferString("42")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to apply uint")
			}
		})

		t.Run("lying loader", func(t *testing.T) {
			var o uint
			s := WithLoader(singleValueLoader(true, Bool, Int))
			if err := Apply(&o, s); !errors.Is(err, ErrLoaderImplementation) {
				t.Error("failed to fail with the right error")
			}
		})
	})
}

func TestApplyToFloat(t *testing.T) {
	t.Run("not a float", func(t *testing.T) {
		var o float64
		y := bytes.NewBufferString("true")
		s := YAML(y)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right value")
		}
	})

	t.Run("not parsed", func(t *testing.T) {
		t.Run("not a number", func(t *testing.T) {
			var o struct{ Foo float64 }
			i := bytes.NewBufferString("foo=true")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right value")
			}
		})

		t.Run("number", func(t *testing.T) {
			var o struct{ Foo float64 }
			i := bytes.NewBufferString("foo=3.14")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != 3.14 {
				t.Error("failed to parse float")
			}
		})
	})

	t.Run("parsed", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			var o float64
			s := WithLoader(singleValueLoader(42, Int, Int|Float))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to parse float")
			}
		})

		t.Run("uint", func(t *testing.T) {
			var o float64
			s := WithLoader(singleValueLoader(uint(42), Int, Int|Float))
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42 {
				t.Error("failed to parse float")
			}
		})

		t.Run("float", func(t *testing.T) {
			var o float64
			y := bytes.NewBufferString("3.14")
			s := YAML(y)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 3.14 {
				t.Error("failed to apply float")
			}
		})

		t.Run("overflow", func(t *testing.T) {
			var o float32
			s := WithLoader(singleValueLoader(math.MaxFloat32 * 2))
			if err := Apply(&o, s); !errors.Is(err, ErrNumericOverflow) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("lying loader", func(t *testing.T) {
			var o float64
			s := WithLoader(singleValueLoader(true, Bool, Float))
			if err := Apply(&o, s); !errors.Is(err, ErrLoaderImplementation) {
				t.Error("failed to fail with the right value", err)
			}
		})
	})
}

func TestApplyToString(t *testing.T) {
	t.Run("not a string", func(t *testing.T) {
		var o string
		j := bytes.NewBufferString("42")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("optional list", func(t *testing.T) {
		t.Run("too many values", func(t *testing.T) {
			var o struct{ Foo string }
			i := bytes.NewBufferString("foo=bar\nfoo=baz")
			s := INI(i)
			if err := Apply(&o, s); !errors.Is(err, ErrTooManyValues) {
				t.Error("failed to fail with the right value")
			}
		})

		t.Run("no values", func(t *testing.T) {
			o := struct{ Foo string }{"hello"}
			i := bytes.NewBufferString("foo.bar=baz\nfoo.bar=qux")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != "hello" {
				t.Error("failed ignore no value")
			}
		})

		t.Run("one value", func(t *testing.T) {
			var o struct{ Foo string }
			i := bytes.NewBufferString("foo=bar")
			s := INI(i)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o.Foo != "bar" {
				t.Error("failed to apply string")
			}
		})
	})

	t.Run("lying loader", func(t *testing.T) {
		var o string
		s := WithLoader(singleValueLoader(42, Int, String))
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("success", func(t *testing.T) {
		var o string
		j := bytes.NewBufferString(`"hello"`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != "hello" {
			t.Error("failed to apply string")
		}
	})
}

func TestApplyToStruct(t *testing.T) {
	t.Run("not a structure", func(t *testing.T) {
		var o struct{ Foo int }
		j := bytes.NewBufferString("42")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("conflicting canonical keys", func(t *testing.T) {
		var o struct{ FooBar int }
		j := bytes.NewBufferString(`{"fooBar": 21, "foo_bar": 42}`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrConflictingKeys) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("non-exported key", func(t *testing.T) {
		o := struct {
			Foo int
			bar int
		}{Foo: 21, bar: 63}
		j := bytes.NewBufferString(`{"foo": 42, "bar": 84}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 || o.bar != 63 {
			t.Error("failed to ignore non-exported field")
		}
	})

	t.Run("non-existing key", func(t *testing.T) {
		o := struct {
			Foo int
			Bar int
		}{Foo: 21, Bar: 63}
		j := bytes.NewBufferString(`{"foo": 42}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 || o.Bar != 63 {
			t.Error("failed to ignore non-existing key")
		}
	})

	t.Run("invalid field value", func(t *testing.T) {
		var o struct{ Foo int }
		j := bytes.NewBufferString(`{"foo": true}`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("apply", func(t *testing.T) {
		var o struct {
			Foo int
			Bar int
		}
		j := bytes.NewBufferString(`{"foo": 42, "bar": 84}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.Foo != 42 || o.Bar != 84 {
			t.Error("failed to apply struct")
		}
	})
}

func TestApplyToMap(t *testing.T) {
	t.Run("non-string key", func(t *testing.T) {
		o := make(map[int]int)
		j := bytes.NewBufferString(`{"21": 42}`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidTarget) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("nil", func(t *testing.T) {
		o := make(map[string]int)
		o["foo"] = 42
		j := bytes.NewBufferString("null")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if _, ok := o["foo"]; ok {
			t.Error("failed to apply map")
		}
	})

	t.Run("not a structure", func(t *testing.T) {
		o := make(map[string]int)
		j := bytes.NewBufferString("42")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("no fields", func(t *testing.T) {
		o := make(map[string]int)
		o["foo"] = 42
		j := bytes.NewBufferString("{}")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o["foo"] != 42 {
			t.Error("failed leave map values")
		}
	})

	t.Run("invalid field type", func(t *testing.T) {
		var o map[string]int
		j := bytes.NewBufferString(`{"foo": "bar"}`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("uninitialized", func(t *testing.T) {
		var o struct{ M map[string]int }
		j := bytes.NewBufferString(`{"m": {"foo": 21, "bar": 42}}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o.M["foo"] != 21 || o.M["bar"] != 42 {
			t.Error("failed to apply map")
			t.Log(o.M == nil)
		}
	})

	t.Run("initialized", func(t *testing.T) {
		o := make(map[string]int)
		o["foo"] = 21
		o["bar"] = 42
		j := bytes.NewBufferString(`{"bar": 63, "baz": 84}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o["foo"] != 21 || o["bar"] != 63 || o["baz"] != 84 {
			t.Error("failed to apply map")
		}
	})
}

func TestApplyToList(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		o := []int{1, 2, 3}
		j := bytes.NewBufferString("null")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 0 {
			t.Error("failed to apply nil to list")
		}
	})

	t.Run("not a list", func(t *testing.T) {
		var o []int
		j := bytes.NewBufferString("42")
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		o := []int{1, 2, 3}
		j := bytes.NewBufferString("[]")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 0 || o == nil {
			t.Error("failed to apply empty list")
		}
	})

	t.Run("invalid item type", func(t *testing.T) {
		var o []int
		j := bytes.NewBufferString(`["bar"]`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("uninitialized", func(t *testing.T) {
		var o []int
		j := bytes.NewBufferString("[1, 2, 3]")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 3 || o[0] != 1 || o[1] != 2 || o[2] != 3 {
			t.Error("failed to apply empty list")
		}
	})

	t.Run("initialized", func(t *testing.T) {
		o := []int{1, 2, 3}
		j := bytes.NewBufferString("[4, 5, 6]")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if len(o) != 3 || o[0] != 4 || o[1] != 5 || o[2] != 6 {
			t.Error("failed to apply empty list")
		}
	})
}

func TestApplyToInterface(t *testing.T) {
	type iface interface{}

	t.Run("nil", func(t *testing.T) {
		t.Run("uninitialized", func(t *testing.T) {
			var o iface
			j := bytes.NewBufferString("null")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != nil {
				t.Error("failed to apply nil")
			}
		})

		t.Run("initialized", func(t *testing.T) {
			var o iface = 42
			j := bytes.NewBufferString("null")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != nil {
				t.Error("failed to apply nil")
			}
		})
	})

	t.Run("primitive", func(t *testing.T) {
		t.Run("does not implement", func(t *testing.T) {
			var o interface{ Foo() }
			j := bytes.NewBufferString("42")
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("uninitialized", func(t *testing.T) {
			var o iface
			j := bytes.NewBufferString("42")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42. {
				t.Error("failed to apply primitive")
			}
		})

		t.Run("initialized", func(t *testing.T) {
			var o iface = "foo"
			j := bytes.NewBufferString("42")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			if o != 42. {
				t.Error("failed to apply primitive")
			}
		})
	})

	t.Run("list", func(t *testing.T) {
		t.Run("does not implement", func(t *testing.T) {
			var o interface{ Foo() }
			j := bytes.NewBufferString("[1, 2, 3]")
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("uninitialized", func(t *testing.T) {
			var o iface
			j := bytes.NewBufferString("[1, 2, 3]")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			l, ok := o.([]interface{})
			if !ok {
				t.Error("failed to apply list")
			}

			if len(l) != 3 || l[0] != 1. || l[1] != 2. || l[2] != 3. {
				t.Error("failed to apply list")
			}
		})

		t.Run("initialized", func(t *testing.T) {
			var o iface = "foo"
			j := bytes.NewBufferString("[1, 2, 3]")
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			l, ok := o.([]interface{})
			if !ok {
				t.Error("failed to apply list")
			}

			if len(l) != 3 || l[0] != 1. || l[1] != 2. || l[2] != 3. {
				t.Error("failed to apply list")
			}
		})
	})

	t.Run("structure", func(t *testing.T) {
		t.Run("does not implement", func(t *testing.T) {
			var o interface{ Foo() }
			j := bytes.NewBufferString(`{"foo": 21, "bar": 42}`)
			s := JSON(j)
			if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
				t.Error("failed to fail with the right error")
			}
		})

		t.Run("uninitialized", func(t *testing.T) {
			var o iface
			j := bytes.NewBufferString(`{"foo": 42, "bar": 84}`)
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			m, ok := o.(map[string]interface{})
			if !ok {
				t.Error("failed to apply structure")
			}

			if len(m) != 2 || m["foo"] != 42. || m["bar"] != 84. {
				t.Error("failed to apply structure")
			}
		})

		t.Run("initialized", func(t *testing.T) {
			var o iface = "foo"
			j := bytes.NewBufferString(`{"foo": 42, "bar": 84}`)
			s := JSON(j)
			if err := Apply(&o, s); err != nil {
				t.Error(err)
			}

			m, ok := o.(map[string]interface{})
			if !ok {
				t.Error("failed to apply structure")
			}

			if len(m) != 2 || m["foo"] != 42. || m["bar"] != 84. {
				t.Error("failed to apply structure")
			}
		})
	})

	t.Run("not implemented source type", func(t *testing.T) {
		var o iface = 21
		s := WithLoader(singleValueLoader(42, Int, ignored))
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if o != 21 {
			t.Error("failed to ignore not implemented source type")
		}
	})
}

func TestApplyToPointer(t *testing.T) {
	t.Run("uninitialized", func(t *testing.T) {
		var o struct{ Foo *int }
		j := bytes.NewBufferString(`{"foo": 42}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if *o.Foo != 42 {
			t.Error("failed to apply to pointer")
		}
	})

	t.Run("elem fails", func(t *testing.T) {
		var o *int
		j := bytes.NewBufferString(`"foo"`)
		s := JSON(j)
		if err := Apply(&o, s); !errors.Is(err, ErrInvalidInputValue) {
			t.Error("failed to fail with the right error")
		}
	})

	t.Run("elem not set", func(t *testing.T) {
		var o struct{ Foo *int }
		i := 42
		o.Foo = &i
		j := bytes.NewBufferString("{}")
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if *o.Foo != 42 {
			t.Error("failed to preserve default value")
		}
	})

	t.Run("elem set", func(t *testing.T) {
		var o struct{ Foo *int }
		i := 21
		o.Foo = &i
		j := bytes.NewBufferString(`{"foo": 42}`)
		s := JSON(j)
		if err := Apply(&o, s); err != nil {
			t.Error(err)
		}

		if *o.Foo != 42 {
			t.Error("failed to preserve default value")
		}
	})
}

func TestApplyToInvalidTarget(t *testing.T) {
	var o func()
	j := bytes.NewBufferString("null")
	s := JSON(j)
	if err := Apply(&o, s); !errors.Is(err, ErrInvalidTarget) {
		t.Error("failed to fail with the right error", err)
	}
}
