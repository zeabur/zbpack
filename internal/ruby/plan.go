package ruby

import (
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultRubyVersion is the default Ruby version of Zeabur.
const DefaultRubyVersion = "3.3"

// DefaultRubyEntrypoint is the default entrypoint of Ruby projects.
const DefaultRubyEntrypoint = "main.rb"

const (
	// ConfigRubyVersion is the configuration key for the Ruby version.
	ConfigRubyVersion = "ruby.version"
	// ConfigRubyEntry is the configuration key for the Ruby entrypoint file.
	//
	// When this configuration is set, the start command will be `ruby <entrypoint>`.
	ConfigRubyEntry = "ruby.entry"
)

// DetermineRubyVersion determines the version of Ruby used in the project.
func DetermineRubyVersion(source afero.Fs, config plan.ImmutableProjectConfiguration) string {
	if version, err := plan.Cast(config.Get(ConfigRubyVersion), cast.ToStringE).Take(); err == nil {
		return version
	}

	reg := regexp.MustCompile(`ruby ["'](\d+\.\d+\.\d+)["']`)
	sourceFile, err := utils.ReadFileToUTF8(source, "Gemfile")
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
	f, err := utils.ReadFileToUTF8(source, "Gemfile")
	if err != nil {
		return types.RubyFrameworkNone
	}

	if strings.Contains(string(f), "rails") {
		return types.RubyFrameworkRails
	}

	return types.RubyFrameworkNone
}

// DetermineBuildCmd determines the build command of the Ruby project.
func DetermineBuildCmd(framework types.RubyFramework, config plan.ImmutableProjectConfiguration) string {
	if cmd, err := plan.Cast(config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		return cmd
	}

	switch framework {
	case types.RubyFrameworkRails:
		return "bundle exec rake assets:precompile"
	}

	return ""
}

// DetermineStartCmd determines the start command of the Ruby project.
func DetermineStartCmd(framework types.RubyFramework, config plan.ImmutableProjectConfiguration) string {
	if cmd, err := plan.Cast(config.Get(plan.ConfigStartCommand), cast.ToStringE).Take(); err == nil {
		return cmd
	}
	entryConfig := plan.Cast(config.Get(ConfigRubyEntry), cast.ToStringE)

	if entryConfig.IsNone() {
		switch framework {
		case types.RubyFrameworkRails:
			return "rails server -b 0.0.0.0 -p 8080"
		}
	}

	return "ruby " + entryConfig.TakeOr(DefaultRubyEntrypoint)
}
