package plan

import (
	"errors"
	"fmt"
	"log"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ImmutableProjectConfiguration declares the common interface for getting values
// in project configuration.
type ImmutableProjectConfiguration interface {
	Get(key string) optional.Option[interface{}]
}

// ProjectConfiguration declares the common interface for project configuration.
type ProjectConfiguration interface {
	ImmutableProjectConfiguration
}

// ViperProjectConfiguration reads the extra configuration "zbpack.toml" from
// the root directory of a project and turns it to a struct for easy access.
type ViperProjectConfiguration struct {
	// root is the configuration for the `zbpack.json`.
	root *viper.Viper
	// submodule is the configuration for the `zbpack.[submodule].json`.
	submodule *viper.Viper
}

func (vpc *ViperProjectConfiguration) Get(key string) optional.Option[interface{}] {
	if vpc.submodule != nil && vpc.submodule.IsSet(key) {
		return optional.Some(vpc.submodule.Get(key))
	}

	if vpc.root != nil && vpc.root.IsSet(key) {
		return optional.Some(vpc.root.Get(key))
	}

	return optional.None[interface{}]()
}

// NewProjectConfigurationFromFs creates a new ViperProjectConfiguration from fs.
//
// If the configuration file is not found, it will print a warning and
// return a default configuration.
func NewProjectConfigurationFromFs(fs afero.Fs, submoduleName string) ProjectConfiguration {
	vpc := &ViperProjectConfiguration{
		root:      nil,
		submodule: nil,
	}

	root, err := loadConfigToViper(fs, "zbpack.json")
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		log.Printf("Failed to read the root configuration file (%s).", err)
	} else {
		vpc.root = root
	}

	if submoduleName != "" {
		submodule, err := loadConfigToViper(fs, fmt.Sprintf("zbpack.%s.json", submoduleName))
		if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
			log.Printf("Failed to read the submodule configuration file (%s).", err)
		} else {
			vpc.submodule = submodule
		}
	}

	return vpc
}

func loadConfigToViper(fs afero.Fs, filename string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("json")

	file, err := fs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", filename, err)
	}
	defer func(file afero.File) {
		err := file.Close()
		if err != nil {
			log.Printf("closing file %s: %s", filename, err)
		}
	}(file)

	if err := v.ReadConfig(file); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", filename, err)
	}

	return v, nil
}

// CastOptionValueOrNone casts the value to the given type.
// If the value is not present or the type assertion fails, it returns None.
func CastOptionValueOrNone[T any](value optional.Option[interface{}], caster func(any) (T, error)) optional.Option[T] {
	innerValue, err := value.Take()
	if err != nil {
		return optional.None[T]()
	}

	if v, ok := innerValue.(T); ok {
		return optional.Some(v)
	}

	// Accept a `cast.To*E()` function.
	cv, err := caster(innerValue)
	if err != nil {
		return optional.None[T]()
	}

	return optional.Some(cv)
}
