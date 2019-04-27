package config

import "testing"

func TestOverride(t *testing.T) {
	t.Skip()

	s := New()
	s.SetFileBase("app/config")
	s.Test.SetFileSystem(testFileSystem{
		"/etc/app/config": `
			foo = 1
			bar = 2
			baz = 3
		`,

		"$HOME/.config/app/config": `
			foo = 4
			bar = 5
			baz = 6
		`,

		"$BINDIR/.config": `
			foo = 7
			bar = 8
			baz = 9
		`,

		".alt-config": `
			foo = 10
			bar = 11
			baz = 12
		`,
	})
	s.Test.SetEnv(map[string]string{
		"FOO": "13",
		"BAR": "14",
	})
	s.SetFileFlagName("config-file")
	s.Test.SetFlags(map[string]string{
		"foo":         "15",
		"config-file": ".alt-config",
	})

	type options struct {
		foo, bar, baz, qux int
	}

	var o options
	o.qux = 16

	if err := s.Apply(&o); err != nil {
		t.Fatal(err)
	}

	if o.foo != 15 {
		t.Error("failed to take the value of 'foo' from the command line flags", o.foo)
	}

	if o.bar != 14 {
		t.Error("failed to take the value of 'bar' from the environment", o.bar)
	}

	if o.baz != 12 {
		t.Error("failed to take the value of 'baz' from the alternative config file", o.baz)
	}

	if o.qux != 16 {
		t.Error("failed to leave the default value of 'qux'", o.qux)
	}
}
