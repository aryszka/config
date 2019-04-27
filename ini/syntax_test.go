package ini

import (
	"bytes"
	"fmt"
	"testing"
)

type (
	result interface{}
	check  func(*Node) result
)

func fail(*Node) result { return "check cannot pass" }
func pass(*Node) result { return nil }

func combine(checks ...check) check {
	return func(n *Node) result {
		for _, c := range checks {
			r := c(n)
			if r == nil || r == true {
				continue
			}

			return r
		}

		return nil
	}
}

func prefix(p string, checks ...check) check {
	return func(n *Node) result {
		r := combine(checks...)(n)
		if r == nil || r == true {
			return r
		}

		return fmt.Sprintf("%s: %v", p, r)
	}
}

func name(value string) check {
	return func(n *Node) result {
		if n.Name == value {
			return nil
		}

		return fmt.Sprintf("invalid name: %s, expected: %s", n.Name, value)
	}
}

func text(value string) check {
	return func(n *Node) result {
		if n.Text() == value {
			return nil
		}

		return fmt.Sprintf("invalid text: %s, expected: %s", n.Text(), value)
	}
}

func length(l int) check {
	return func(n *Node) result {
		if len(n.Nodes) == l {
			return nil
		}

		return fmt.Sprintf("invalid length: %d, expected: %d", len(n.Nodes), l)
	}
}

func children(c ...check) check {
	return combine(length(len(c)), func(n *Node) result {
		for i := range c {
			r := c[i](n.Nodes[i])
			if r == nil || r == true {
				continue
			}

			return fmt.Sprintf("invalid child at %d: %v", i, r)
		}

		return nil
	})
}

func value(txt string) check {
	return prefix("value", name("value"), text(txt))
}

func stringsToChecks(m func(string) check, s ...string) []check {
	var checks []check
	for i := range s {
		checks = append(checks, m(s[i]))
	}

	return checks
}

func values(texts ...string) check {
	return children(stringsToChecks(value, texts...)...)
}

func symbol(s string) check {
	return prefix("symbol", name("symbol"), text(s))
}

func key(symbols ...string) check {
	return prefix("key", name("key"), children(stringsToChecks(symbol, symbols...)...))
}

func keyedValue(parts ...string) check {
	if len(parts) < 2 {
		return fail
	}

	return prefix(
		"keyed-value",
		name("keyed-value"),
		children(
			key(parts[:len(parts)-1]...),
			value(parts[len(parts)-1]),
		),
	)
}

func groupKey(keyCheck check) check {
	return prefix("group-key", name("group-key"), children(keyCheck))
}

func group(keyCheck check, valueChecks ...check) check {
	return prefix("group", children(append([]check{groupKey(keyCheck)}, valueChecks...)...))
}

func testFail(t *testing.T, title string, configs ...string) {
	for _, c := range configs {
		t.Run(title, func(t *testing.T) {
			if _, err := parse(bytes.NewBufferString(c)); err == nil {
				t.Error("failed to fail")
			}
		})
	}
}

func testSucceed(t *testing.T, title, config string, checks ...check) {
	t.Run(title, func(t *testing.T) {
		n, err := parse(bytes.NewBufferString(config))
		if err != nil {
			t.Fatal(err)
		}

		result := combine(checks...)(n)
		if result == nil || result == true {
			return
		}

		t.Error(result)
	})
}

func testEscapedChars(t *testing.T, configs ...string) {
	for _, c := range configs {
		testSucceed(t, "escaped char", c, children(name("keyed-value")))
	}
}

func TestParseIni(t *testing.T) {
	testSucceed(t, "empty", "", length(0))
	testSucceed(t, "comment only", "# foo", length(0))

	testFail(t, "invalid key chars", "+ = 42")
	testFail(t, "empty key", "= 42")
	testFail(t, "missing value", "foo =")
	testSucceed(t, "keyed value", "foo = 42", children(keyedValue("foo", "42")))
	testFail(t, "unclosed quote", "foo = \"bar")
	testFail(t, "unclosed single quote", "foo = 'bar")
	testFail(t, "invalid chars", "foo = \\", "foo = [", "foo = ]", "foo = =")
	testEscapedChars(t, "foo = \\\\", "foo = \\[", "foo = \\]", "foo = \\=")
	testSucceed(t, "keyed, quoted value", "foo = \"bar\"", children(combine(
		keyedValue("foo", "\"bar\""),
		children(pass, children(name("quote"))),
	)))
	testSucceed(t, "keyed, single quoted value", "foo = 'bar'", children(combine(
		keyedValue("foo", "'bar'"),
		children(pass, children(name("quote"))),
	)))
	testFail(t, "cannot have comments in a key", "foo#bar = 42")
	testSucceed(t, "comment after a keyed value", "foo = 42 # bar", children(keyedValue("foo", "42")))

	testFail(t, "invalid composite key", "foo. = 42", "bar:: = 42")
	testSucceed(t, "composite key, dot", "foo.bar = 42", children(keyedValue("foo", "bar", "42")))
	testSucceed(t, "composite key, colons", "foo::bar = 42", children(keyedValue("foo", "bar", "42")))
	testSucceed(t, "composite key, trim start", " foo.bar = 42", children(keyedValue("foo", "bar", "42")))

	testFail(t, "invalid group key", "[]", "[+]", "[foo bar]")
	testSucceed(t, "empty group", "[foo]", children(group(key("foo"))))
	testSucceed(t, "group key trimmed", "[ foo ]", children(group(key("foo"))))
	testSucceed(t, "composite group key", "[foo.bar.baz]", children(group(key("foo", "bar", "baz"))))

	const groupWithSimpleValues = `
		[foo.bar.baz]
		1
		2
		3
	`
	testSucceed(t, "group with simple values", groupWithSimpleValues, children(group(
		key("foo", "bar", "baz"),
		value("1"),
		value("2"),
		value("3"),
	)))

	const groupWithKeyedValues = `
		[foo.bar.baz]
		a = 1
		b = 2
		c = 3
	`
	testSucceed(t, "group with keyed values", groupWithKeyedValues, children(group(
		key("foo", "bar", "baz"),
		keyedValue("a", "1"),
		keyedValue("b", "2"),
		keyedValue("c", "3"),
	)))

	const groupWithMixedValues = `
		[foo.bar.baz]
		42
		a = 1
	`
	testSucceed(t, "group with mixed values", groupWithMixedValues, children(group(
		key("foo", "bar", "baz"),
		value("42"),
		keyedValue("a", "1"),
	)))

	const commentAfterAGroupKey = `
		[foo.bar.baz] # qux
		1
		2
		3
	`
	testSucceed(t, "comment after a group key", commentAfterAGroupKey, children(group(
		key("foo", "bar", "baz"),
		value("1"),
		value("2"),
		value("3"),
	)))

	const groupTerminated = `
		[foo]
		a = 1
		b = 2

		c = 3
	`
	testSucceed(t, "group terminated", groupTerminated, children(
		group(
			key("foo"),
			keyedValue("a", "1"),
			keyedValue("b", "2"),
		),
		keyedValue("c", "3"),
	))

	const groupWithComment = `
		[foo]
		1
		2
		# 3
		4
	`
	testSucceed(t, "group with comment", groupWithComment, children(group(
		key("foo"),
		value("1"),
		value("2"),
		value("4"),
	)))

	const multipleGroups = `
		[foo]
		1
		2
		3
		[bar]
		4
		5
		6
		[baz]
		7
		8
		9
	`
	testSucceed(t, "multiple groups", multipleGroups, children(
		group(
			key("foo"),
			value("1"),
			value("2"),
			value("3"),
		),
		group(
			key("bar"),
			value("4"),
			value("5"),
			value("6"),
		),
		group(
			key("baz"),
			value("7"),
			value("8"),
			value("9"),
		),
	))

	const mixedConfig = `
		# This is a config file

		foo = 42
		bar = https://example.org

		[application.secrets]
		user = ./user.json
		client = "./client.json"

		# This is some documentation:
		# hey!

		baz = 84
	`
	testSucceed(t, "mixed config", mixedConfig, children(
		keyedValue("foo", "42"),
		keyedValue("bar", "https://example.org"),
		group(
			key("application", "secrets"),
			keyedValue("user", "./user.json"),
			combine(
				keyedValue("client", "\"./client.json\""),
				children(
					pass,
					children(name("quote")),
				),
			),
		),
		keyedValue("baz", "84"),
	))
}
