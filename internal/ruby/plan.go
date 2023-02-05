package ruby

import (
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineRubyVersion(absPath string) string {
	RubyVersion := GemfileParser(absPath, `ruby "`)

	return RubyVersion
}
func DetermineRubyFramework(absPath string) RubyFramework {
	railsVersion := GemfileParser(absPath, `rails`)
	if railsVersion != "" {
		return RubyFrameworkRails
	}
	return RubyFrameworkNone
}
