package ruby

import (
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
	"regexp"
	"strings"
)

// DetermineRubyVersion determines the version of Ruby used in the project.
func DetermineRubyVersion(source afero.Fs) string {
	reg := regexp.MustCompile(`ruby "(\d+\.\d+\.\d+)"`)
	sourceFile, err := afero.ReadFile(source, "Gemfile")
	if err != nil {
		return ""
	}

	matches := reg.FindStringSubmatch(string(sourceFile))
	if len(matches) < 2 {
		return ""
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
