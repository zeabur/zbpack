package ruby

import (
	"github.com/spf13/afero"
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineRubyVersion(source afero.Fs) string {
	RubyVersion := GetGemfileValue(source, `ruby "`)

	return RubyVersion
}

func DetermineRubyFramework(source afero.Fs) RubyFramework {
	railsVersion := GetGemfileValue(source, `rails`)
	if railsVersion != "" {
		return RubyFrameworkRails
	}
	return RubyFrameworkNone
}
