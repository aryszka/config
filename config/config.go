package config

// TODO:
// - how to deal with the annotations for yaml and json?
// - how to control file types
// - how to handle unsupported keys in files
// - how to handle which possible file sources are allowed
// - support env prefix
// - with flag parsing, we need to know the type in advance, because of the last non-positional bool flag
// doesn't necessarily have a value

type Test struct {
	fileSystem map[string]string
	env        map[string]string
	flags      map[string]string
}

type Settings struct {
	base         string
	fileFlagName string
	Test         *Test
}

func (t *Test) SetFileSystem(fs map[string]string) {
	t.fileSystem = fs
}

func (t *Test) SetEnv(e map[string]string) {
	t.env = e
}

func (t *Test) SetFlags(f map[string]string) {
	t.flags = f
}

func New() *Settings {
	return &Settings{
		Test: &Test{},
	}
}

func (s *Settings) SetBasePath(base string) {
	s.base = base
}

func (s *Settings) SetFileFlagName(n string) {
	s.fileFlagName = n
}

func (s *Settings) Apply(o interface{}) error {
	return nil
}
