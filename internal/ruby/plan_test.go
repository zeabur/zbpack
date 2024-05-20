package ruby_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestDetermineRubyVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Gemfile", []byte(`ruby "2.7.2"`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")

	version := ruby.DetermineRubyVersion(fs, config)
	assert.Equal(t, "2.7.2", version)
}

func TestDetermineRubyVersion_Customized(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(ruby.ConfigRubyVersion, "12.34.56")

	version := ruby.DetermineRubyVersion(fs, config)
	assert.Equal(t, "12.34.56", version)
}

func TestDetermineRubyVersion_Default(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")

	version := ruby.DetermineRubyVersion(fs, config)
	assert.Equal(t, ruby.DefaultRubyVersion, version)
}

func TestDetermineBuildCmd_RailsCallsRake(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	buildCmd := ruby.DetermineBuildCmd(types.RubyFrameworkRails, config)

	assert.Contains(t, buildCmd, "rake assets:precompile")
}

func TestDetermineBuildCmd_NoneReturnsNone(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	buildCmd := ruby.DetermineBuildCmd(types.RubyFrameworkNone, config)

	assert.Equal(t, "", buildCmd)
}

func TestDetermineRubyFramework_Custom(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigBuildCommand, "echo 'hi'")

	buildCmd := ruby.DetermineBuildCmd(types.RubyFrameworkNone, config)
	assert.Equal(t, "echo 'hi'", buildCmd)

	buildCmd = ruby.DetermineBuildCmd(types.RubyFrameworkRails, config)
	assert.Equal(t, "echo 'hi'", buildCmd)
}

func TestDetermineStartCmd_Rails(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	startCmd := ruby.DetermineStartCmd(types.RubyFrameworkRails, config)

	assert.Contains(t, startCmd, "rails server")
}

func TestDetermineStartCmd_None(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, ruby.DefaultRubyEntrypoint, []byte(`puts "Hello, world!"`), 0o644)
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	startCmd := ruby.DetermineStartCmd(types.RubyFrameworkNone, config)

	// default entrypoint
	assert.Equal(t, "ruby "+ruby.DefaultRubyEntrypoint, startCmd)
}

func TestDetermineStartCmd_Custom(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigStartCommand, "echo 'hi'")

	startCmd := ruby.DetermineStartCmd(types.RubyFrameworkNone, config)
	assert.Equal(t, "echo 'hi'", startCmd)

	startCmd = ruby.DetermineStartCmd(types.RubyFrameworkRails, config)
	assert.Equal(t, "echo 'hi'", startCmd)
}

func TestDetermineStartCmd_CustomEntry(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "app.rb", []byte(`puts "Hello, world!"`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(ruby.ConfigRubyEntry, "app.rb")

	startCmd := ruby.DetermineStartCmd(types.RubyFrameworkNone, config)
	assert.Equal(t, "ruby app.rb", startCmd)

	startCmd = ruby.DetermineStartCmd(types.RubyFrameworkRails, config)
	assert.Equal(t, "ruby app.rb", startCmd)
}
