package ruby

import (
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineRubyVersion(absPath string) string {
	RubyVersion := GetGemfileValue(absPath, `ruby "`)

	return RubyVersion
}
func DetermineRubyFramework(absPath string) RubyFramework {
	railsVersion := GetGemfileValue(absPath, `rails`)
	if railsVersion != "" {
		return RubyFrameworkRails
	}
	return RubyFrameworkNone
}
