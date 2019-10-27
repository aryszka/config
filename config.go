package config

/*
import "github.com/aryszka/config/configsource"

func apply(target interface{}, withPositional bool) ([]string, error) {
	s, err := configsource.Default()
	if err != nil {
		return nil, err
	}

	if err := configsource.Apply(s, target); err != nil {
		return nil, err
	}

	if !withPositional {
		return nil, nil
	}

	return configsource.Positional(s), nil
}

func Apply(target interface{}) error {
	_, err := apply(target, false)
	return err
}

func WithPositional(target interface{}) ([]string, error) {
	return apply(target, true)
}
*/
