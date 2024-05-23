package plan

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/iancoleman/strcase"
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ImmutableProjectConfiguration declares the common interface for getting values
// in project configuration.
type ImmutableProjectConfiguration interface {
	// Get returns the value of the given key. If the key is not present, it returns None.
	Get(key string) optional.Option[any]
}

// MutableProjectConfiguration declares the common interface for setting values
// in project configuration.
type MutableProjectConfiguration interface {
	// Set sets the value of the given key. The value set here has the highest priority.
	Set(key string, val any)
}

// ProjectConfiguration declares the common interface for project configuration.
type ProjectConfiguration interface {
	ImmutableProjectConfiguration
	MutableProjectConfiguration
}

// ViperProjectConfiguration reads the extra configuration from the environment
// variable "ZBPACK_[CONFIG_KEY]" and "zbpack.toml" in the root directory of
// a project and turns it to a struct for easy access.
type ViperProjectConfiguration struct {
	// root is the configuration for the `zbpack.json`.
	root *viper.Viper
	// submodule is the configuration for the `zbpack.[submodule].json`.
	submodule *viper.Viper
	// extra is the manual overridden value of this configuration.
	extra map[string]any
}

// Get returns the value of the given key. If the key is not present, it returns None.
func (vpc *ViperProjectConfiguration) Get(key string) optional.Option[any] {
	/* extra */

	if val, ok := vpc.extra[key]; ok {
		return optional.Some(val)
	}

	/* env */

	// FORCE_CONTAINERIZED {"serverless": false}
	if key == "serverless" {
		if v, err := ToWeakBoolE(os.Getenv("FORCE_CONTAINERIZED")); err == nil && v {
			return optional.Some[any](false)
		}
	}

	// ZOLA_VERSION {"zola_version: "1.2.3"}
	if key == "zolaVersion" || key == "zola_version" {
		if val, ok := os.LookupEnv("ZOLA_VERSION"); ok {
			return optional.Some[any](val)
		}
	}

	// key.a.b.c -> ZBPACK_KEY_A_B_C
	envKey := "ZBPACK_" + strcase.ToScreamingSnake(key)
	if val, ok := os.LookupEnv(envKey); ok {
		return optional.Some[any](val)
	}

	/* zbpack.json */

	if vpc.submodule != nil && vpc.submodule.IsSet(key) {
		return optional.Some(vpc.submodule.Get(key))
	}

	if vpc.root != nil && vpc.root.IsSet(key) {
		return optional.Some(vpc.root.Get(key))
	}

	return optional.None[any]()
}

// Set sets the value of the given key. The value set here has the highest priority.
func (vpc *ViperProjectConfiguration) Set(key string, val any) {
	if vpc.extra == nil {
		vpc.extra = make(map[string]any)
	}

	vpc.extra[key] = val
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

// Cast casts the value to the given type.
// If the value is not present or the type assertion fails, it returns None.
func Cast[T any](value optional.Option[any], caster func(any) (T, error)) optional.Option[T] {
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

// Common configuration keys.
const (
	// ConfigInstallCommand is the key for the installation command in the project configuration.
	ConfigInstallCommand = "install_command"
	// ConfigBuildCommand is the key for the build command in the project configuration.
	ConfigBuildCommand = "build_command"
	// ConfigStartCommand is the key for the start command in the project configuration.
	ConfigStartCommand = "start_command"
	// ConfigOutputDir is the key for the output directory in the project configuration.
	ConfigOutputDir = "output_dir"
)
