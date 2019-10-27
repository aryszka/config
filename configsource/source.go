package configsource

/*
import (
	"errors"
	"fmt"
	"strings"
)

const customConfigName = "config" // TODO: support the -c shortcut

type SourceFlags int

const (
	None      SourceFlags = 0
	Primitive SourceFlags = 1 << iota
	List
	Struct
	Unparsed
	Flaky
)

type Source interface {
	Flags() SourceFlags
	Len() int
	Value(int) interface{}
	Fields() []string
	HasField(string) bool
	Field(string) Source
}

type Loader interface {
	Load() (Source, error)

	/*
		// TODO: these don't necessarily belong here
		Abbreviations(interface{})
		Undefined() [][]string
*/ /*
}

type ConfigTarget interface {
	Apply(Source) error
}

type delayedError struct {
	err error
}

var NotFound = errors.New("not found")

var notFoundLoader = delayedError{NotFound}

func (f SourceFlags) has(flag SourceFlags) bool {
	return f&flag != 0
}

func (f SourceFlags) kindIncludes(kind SourceFlags) bool {
	return f&(Primitive|List|Struct) == 0 || f.has(kind)
}

func (f SourceFlags) String() string {
	switch f {
	case None:
		return "none"
	case Primitive:
		return "primitive"
	case List:
		return "list"
	case Struct:
		return "struct"
	case Unparsed:
		return "unparsed"
	case Flaky:
		return "flaky"
	}

	var s []string
	for v := Primitive; v <= Flaky; v <<= 1 {
		if f.has(v) {
			s = append(s, v.String())
		}
	}

	return strings.Join(s, "/")
}

func (d delayedError) Load() (Source, error) {
	return nil, d.err
}

func Default() (Source, error) {
	appName := getAppName()
	home := getHome()
	binDir := getBinDir()
	envLoader := Once(Env(EnvOptions{Prefix: appName}))
	flagLoader := Once(Flags(FlagOptions{}))
	loader := Merge(
		Override(
			allFormats(fmt.Sprintf("/etc/%s/config", appName)),
			allFormats(fmt.Sprintf("%s/.config/%s/config", home, appName)),
			allFormats(fmt.Sprintf("%s/.%s", binDir, appName)),
			customConfigFile(customConfigName, Override(envLoader, flagLoader)),
		),
		envLoader,
		flagLoader,
	)

	return loader.Load()
}

func Apply(s Source, target interface{}) error {
	err := applySource(s, target)
	if err == NotFound {
		err = nil
	}

	return err
}

func Positional(s Source) []string {
	var p []string
	for i := 0; i < s.Len(); i++ {
		p = append(p, fmt.Sprint(s.Value(i)))
	}

	return p
}

// TODO:
// - maybe the help is not needed at all
// - what to do with file errors, e.g. no permission? ApplySafe()? Or consider EPERM as NotFound? Or just live
// with it?
// - error reporting: undefined fields
// - how to make this testable?
// - provide options for reporting unused
// - allow source specifiec control over undefined entries. Use the term 'unexpected'
// - need support for help. Usage from both annotations and a provided map
// - support a help flag
// - how to support subcommands? Here rather no
// - reconsider the file override/merge use case
// - rename to options?
*/
