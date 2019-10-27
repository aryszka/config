package configsource

/*
import (
	"io/ioutil"
	"os"
)

type file struct {
	name string
}

func allFormats(name string) Loader {
	return Override(
		YAML(name+".yml"),
		YAML(name+".yaml"),
		JSON(name+".json"),
		TOML(name+".toml"),
		INI(name+".ini"),
	)
}

func loadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if os.IsNotExist(err) {
		return nil, NotFound
	}

	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func customConfigFile(name string, l Loader) Loader {
	s, err := l.Load()
	if err != nil {
		return delayedError{err}
	}

	if !s.HasField(name) {
		return delayedError{NotFound}
	}

	f := s.Field(name)
	var fn string
	if err := Apply(f, &fn); err != nil {
		return delayedError{err}
	}

	if fn == "" {
		return delayedError{NotFound}
	}

	return ByExtension(fn)
}

func ByExtension(name string) Loader {
	return nil
}
*/
