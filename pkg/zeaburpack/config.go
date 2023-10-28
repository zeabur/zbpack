package zeaburpack

import (
	"reflect"

	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/pkg/plan"
)

// ZbpackOptions is the union type of the pointer to PlanOptions and BuildOptions.
type ZbpackOptions interface {
	*PlanOptions | *BuildOptions
}

// UpdateOptionsOnConfig updates the undefined options (nil) with the value from project configuration.
func UpdateOptionsOnConfig[T ZbpackOptions](options T, config plan.ImmutableProjectConfiguration) {
	// Reflect to put CustomBuildCommand, CustomStartCommand, and OutputDir
	structReflection := reflect.Indirect(reflect.ValueOf(options))
	customBuildCommand := structReflection.FieldByName("CustomBuildCommand")
	customStartCommand := structReflection.FieldByName("CustomStartCommand")
	outputDir := structReflection.FieldByName("OutputDir")

	if !customBuildCommand.IsValid() || !customStartCommand.IsValid() || !outputDir.IsValid() {
		panic("Not a valid ZbpackOptions.")
	}

	// You can specify customBuildCommand, customStartCommand, and
	// outputDir in the project configuration file, with the following
	// form:
	//
	// {"build_command": "your_command"}
	// {"start_command": "your_command"}
	// {"output_dir": "your_output_dir"}
	//
	// The submodule-specific configuration (zbpack.[submodule].json)
	// overrides the project configuration if defined.
	if customBuildCommand.IsNil() {
		value, err := plan.Cast(config.Get("build_command"), cast.ToStringE).Take()
		if err == nil {
			customBuildCommand.Set(reflect.ValueOf(&value))
		}
	}
	if customStartCommand.IsNil() {
		value, err := plan.Cast(config.Get("start_command"), cast.ToStringE).Take()
		if err == nil {
			customStartCommand.Set(reflect.ValueOf(&value))
		}
	}
	if outputDir.IsNil() {
		value, err := plan.Cast(config.Get("output_dir"), cast.ToStringE).Take()
		if err == nil {
			outputDir.Set(reflect.ValueOf(&value))
		}
	}
}
