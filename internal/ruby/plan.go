package ruby

import (
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// DetermineRubyVersion determines the version of Ruby used in the project.
func DetermineRubyVersion(source afero.Fs) string {
	RubyVersion := GetGemfileValue(source, `ruby "`)

	return RubyVersion
}

// DetermineRubyFramework determines the framework of the Ruby project.
func DetermineRubyFramework(source afero.Fs) types.RubyFramework {
	railsVersion := GetGemfileValue(source, `rails`)
	if railsVersion != "" {
		return types.RubyFrameworkRails
	}
	return types.RubyFrameworkNone
}
