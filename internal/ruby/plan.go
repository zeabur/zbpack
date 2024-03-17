package ruby

import (
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultRubyVersion is the default Ruby version of Zeabur.
const DefaultRubyVersion = "3.3"

// DetermineRubyVersion determines the version of Ruby used in the project.
func DetermineRubyVersion(source afero.Fs) string {
	reg := regexp.MustCompile(`ruby ["'](\d+\.\d+\.\d+)["']`)
	sourceFile, err := afero.ReadFile(source, "Gemfile")
	if err != nil {
		return DefaultRubyVersion
	}

	matches := reg.FindStringSubmatch(string(sourceFile))
	if len(matches) < 2 {
		return DefaultRubyVersion
	}

	return matches[1]
}

// DetermineRubyFramework determines the framework of the Ruby project.
func DetermineRubyFramework(source afero.Fs) types.RubyFramework {
	f, err := afero.ReadFile(source, "Gemfile")
	if err != nil {
		return types.RubyFrameworkNone
	}

	if strings.Contains(string(f), "rails") {
		return types.RubyFrameworkRails
	}

	return types.RubyFrameworkNone
}
