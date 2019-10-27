package configsource

/*
import (
	"fmt"
	"os"
	"strings"
)

type FlagsMode int

const (
	FlagsDefault FlagsMode = 0
	SingleDash   FlagsMode = 1 << iota
	BanShort
	BanGroupedShort
	PositionalLast
)

type FlagOptions struct {
	Args *[]string
	Mode FlagsMode
}

type flag struct {
	name, value string
	key         []string
	hasValue    bool
}

type flags struct {
	options FlagOptions
}

func errFlagAfterPositional(name string) error {
	return fmt.Errorf("flag after positional argument not allowed: %s", name)
}

func errDoubleDashNotAllowed(name string) error {
	return fmt.Errorf("double dash not allowed: %s", name)
}

func errShortNotAllowed(name string) error {
	return fmt.Errorf("short flag not allowed: %s", name)
}

func errGroupedShortNotAllowed(name string) error {
	return fmt.Errorf("flag grouping not allowed: %s", name)
}

func Flags(o FlagOptions) Loader {
	if o.Args == nil {
		args := make([]string, len(os.Args)-1)
		copy(args, os.Args[1:])
		o.Args = &args
	}

	return flags{options: o}
}

func groupFlags(m FlagsMode, args []string) (f []flag, p []string, err error) {
	positionalLast := m&PositionalLast != 0
	var hasPositional bool
	for i := 0; i < len(args); i++ {
		current := args[i]
		if current == "--" {
			p = append(p, args[i+1:]...)
			return
		}

		if !strings.HasPrefix(current, "-") || len(current) == 1 {
			p = append(p, current)
			hasPositional = true
			continue
		}

		if positionalLast && hasPositional {
			err = errFlagAfterPositional(current)
			return
		}

		if i == len(args)-1 {
			f = append(f, flag{name: current})
			return
		}

		next := args[i+1]
		if next == "--" || strings.HasPrefix(next, "-") && len(next) > 1 {
			f = append(f, flag{name: current})
			continue
		}

		f = append(f, flag{name: current, value: next, hasValue: true})
		i++
	}

	return
}

func processKeys(m FlagsMode, f []flag) ([]flag, error) {
	var r []flag
	singleDashMode := m&SingleDash != 0
	banShort := m&BanShort != 0
	banGroupedShort := m&BanGroupedShort != 0
	for _, fi := range f {
		hasDoubleDash := strings.HasPrefix(fi.name, "--")
		if hasDoubleDash && singleDashMode {
			return nil, errDoubleDashNotAllowed(fi.name)
		}

		if hasDoubleDash {
			fi.key = strings.Split(fi.name[2:], ".")
			r = append(r, fi)
			continue
		}

		if singleDashMode {
			fi.key = strings.Split(fi.name[1:], ".")
			r = append(r, fi)
			continue
		}

		if banShort {
			return nil, errShortNotAllowed(fi.name)
		}

		if banGroupedShort && len(fi.name) > 2 {
			return nil, errGroupedShortNotAllowed(fi.name)
		}

		for _, symbol := range strings.Split(fi.name[1:len(fi.name)-1], "") {
			r = append(r, flag{name: fi.name, key: []string{symbol}})
		}

		fi.key = []string{fi.name[len(fi.name)-1:]}
		r = append(r, fi)
	}

	return r, nil
}

func addFlag(n node, f flag) node {
	if len(f.key) == 0 {
		v := f.value
		if !f.hasValue {
			v = "true"
		}

		n.flags |= Unparsed
		n.values = append(n.values, v)
		return n
	}

	symbol := f.key[0]
	f.key = f.key[1:]
	if n.fields == nil {
		n.fields = make(map[string]Source)
	}

	nf, _ := n.fields[symbol].(node)
	n.fields[symbol] = addFlag(nf, f)
	return n
}

func buildSource(f []flag, p []string) Source {
	n := node{flags: Unparsed}
	for _, fi := range f {
		n = addFlag(n, fi)
	}

	for _, pi := range p {
		n.values = append(n.values, pi)
	}

	return n
}

func (f flags) Load() (Source, error) {
	fs, p, err := groupFlags(f.options.Mode, *f.options.Args)
	if err != nil {
		return nil, err
	}

	if fs, err = processKeys(f.options.Mode, fs); err != nil {
		return nil, err
	}

	return buildSource(fs, p), nil
}
*/
