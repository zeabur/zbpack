package zeaburpack

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ImmutableProjectConfiguration declares the common interface for getting values
// in project configuration.
type ImmutableProjectConfiguration interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetIntSlice(key string) []int
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetUint(key string) uint
	GetUint16(key string) uint16
	GetUint32(key string) uint32
	GetUint64(key string) uint64
}

// ProjectConfiguration declares the common interface for project configuration.
type ProjectConfiguration interface {
	ImmutableProjectConfiguration
}

// ViperProjectConfiguration reads the extra configuration "zbpack.toml" from
// the root directory of project and turn it to a struct for easy access.
type ViperProjectConfiguration struct {
	*viper.Viper
}

// NewProjectConfiguration creates a new ViperProjectConfiguration.
func NewProjectConfiguration() ProjectConfiguration {
	viper.SetConfigType("toml")

	return &ViperProjectConfiguration{
		Viper: viper.New(),
	}
}

// NewProjectConfigurationFromFs creates a new ViperProjectConfiguration from fs.
//
// If the configuration file is not found, it will print a warning and
// return a default configuration.
func NewProjectConfigurationFromFs(fs afero.Fs) ProjectConfiguration {
	vpc := NewProjectConfiguration().(*ViperProjectConfiguration)
	err := vpc.ReadFromFs(fs)
	if err != nil {
		log.Println("read config from fs:", err)
		return NewProjectConfiguration()
	}

	return vpc
}

// ReadFromFs reads the configuration from the given file system.
func (vpc *ViperProjectConfiguration) ReadFromFs(fs afero.Fs) error {
	file, err := fs.Open("zbpack.toml")
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer func(file afero.File) {
		err := file.Close()
		if err != nil {
			log.Println("closing file:", err)
		}
	}(file)

	return vpc.ReadConfig(file)
}
