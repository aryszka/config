# INI Syntax

A key and a value may look like this:

```
foo = 42
```

A key can consist of multiple symbols:

```
foo.bar.baz = 42
```

In this case, when the config is processed, the value 42 will be mapped to the baz field of a structure found
on the bar field of a structure found on the foo field of the root structure that is used for parsing.

Entries that look as the following:

```
foo.bar.baz = 42
foo.bar.qux = 84
```

...can be simplified by grouping:

```
[foo.bar]
baz = 42
qux = 84
```

Groups are terminated by an empty line or by another group and cannot be nested.

Defining multiple values for the same field (a form of listing):

```
foo.bar.baz = 1
foo.bar.baz = 2
foo.bar.baz = 3
```

This can be simplified by using a group:

```
[foo.bar.baz]
1
2
3
```

Structured or nested lists are not supported by the config format.

Concepts in the syntax:

- **comment:**
  starts with a # character and is terminated by a \n.
- **symbol:**
  used for mapping config values to fields in an in-memory structure. Can contain the following characters:
  _-a-zA-Z0-9.
- **key:**
  contains one or more symbols, separated by . or ::, without whitespace. It's used as the path to find the
  right field in an in-memory structure.
- **value:**
  can contain any characters except for \\, ", ', [, ], =, \\n, #. When some of these characters are required in a
  value, then escaping can be used with \, or the value can be quoted with " or '. When a value is quoted with
  ", then any character can be used except for \\ and ", which can be escaped with \\. When a value is quoted with
  ', then any character can be used except for \\ and ', which can be escaped with \\.
- **keyed value:**
  consists of a key and a value separated by a = character.
- **group:**
  used for prefixing the following keyed values or values with a common key. It's defined by a key between [ and
  ]. The group is terminated by a double newline or another group.
- **whitespace:**
  the syntax ignores whitespace, except for requiring a newline as a separator. Keys are not allowed to have
  whitespace between the contained symbols.
