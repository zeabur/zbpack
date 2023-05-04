package ruby

import (
	"github.com/zeabur/zbpack/internal/source"
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineRubyVersion(source *source.Source) string {
	RubyVersion := GetGemfileValue(source, `ruby "`)

	return RubyVersion
}
func DetermineRubyFramework(source *source.Source) RubyFramework {
	railsVersion := GetGemfileValue(source, `rails`)
	if railsVersion != "" {
		return RubyFrameworkRails
	}
	return RubyFrameworkNone
}
