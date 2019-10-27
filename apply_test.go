package config

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

type testLoader struct {
	value interface{}
	fail  bool
	lie   map[NodeType]NodeType
}

var errTestLoadFailed = errors.New("test load failed")

func singleValueLoader(value interface{}) Loader {
	return testLoader{value: value}
}

func failingLoader() Loader {
	return testLoader{fail: true}
}

func lyingLoader(value interface{}, from, to NodeType) Loader {
	return testLoader{value: value, lie: map[NodeType]NodeType{from: to}}
}

func (l testLoader) Load() (interface{}, error) {
	if l.fail {
		return nil, errTestLoadFailed
	}

	return l.value, nil
}

func (l testLoader) TypeMapping() map[NodeType]NodeType {
	return l.lie
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
		s := WithLoader(lyingLoader(42, Int, Bool))
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
			s := WithLoader(lyingLoader(true, Bool, Int))
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
			s := WithLoader(lyingLoader(true, Bool, Int))
			if err := Apply(&o, s); !errors.Is(err, ErrLoaderImplementation) {
				t.Error("failed to fail with the right error")
			}
		})
	})
}

func TestApplyToInterface(t *testing.T) {
	type iface interface{}

	t.Run("nil", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
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
		t.Run("nil", func(t *testing.T) {
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
		t.Run("nil", func(t *testing.T) {
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
		t.Run("nil", func(t *testing.T) {
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
}
