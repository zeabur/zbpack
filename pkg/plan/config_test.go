package plan_test

import (
	"testing"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestProjectConfiguration_Empty(t *testing.T) {
	t.Parallel()

	config := plan.NewProjectConfiguration()

	assert.Equal(t, "", config.GetString("laravel.test"))
	assert.False(t, config.IsSet("laravel.owo"))
}

func TestProjectConfiguration_ZbpackTomlExisted(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[laravel]
test = "owo"`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs)

	assert.Equal(t, "owo", config.GetString("laravel.test"))
	assert.True(t, config.IsSet("laravel"))
	assert.True(t, config.IsSet("laravel.test"))
	assert.False(t, config.IsSet("laravel.owo"))
}

func TestProjectConfiguration_ZbpackTomlNotExisted(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()

	config := plan.NewProjectConfigurationFromFs(fs)

	assert.Equal(t, "", config.GetString("laravel.test"))
	assert.False(t, config.IsSet("laravel"))
	assert.False(t, config.IsSet("laravel.test"))
	assert.False(t, config.IsSet("laravel.owo"))
}

func TestGetProjectConfigValue_Global(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project]
build_command = "build"`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs)
	assert.Equal(t, optional.Some("build"), plan.GetProjectConfigValue(config, "", "build_command"))
}

func TestGetProjectConfigValue_Submodule(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project.sm]
build_command = "build.sm"`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs)
	assert.Equal(t, optional.Some("build.sm"), plan.GetProjectConfigValue(config, "sm", "build_command"))
}

func TestGetProjectConfigValue_SubmoduleOverride(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project]
build_command = "build"
[project.sm]
build_command = "build.sm"`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs)
	assert.Equal(t, optional.Some("build.sm"), plan.GetProjectConfigValue(config, "sm", "build_command"))
}

func TestGetProjectConfigValue_SubmoduleFallback(t *testing.T) {
	t.Parallel()

	t.Run("with-submodule-group", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project]
build_command = "build"
[project.sm]`), 0644)

		config := plan.NewProjectConfigurationFromFs(fs)
		assert.Equal(t, optional.Some("build"), plan.GetProjectConfigValue(config, "sm", "build_command"))
	})

	t.Run("without-submodule-group", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project]
build_command = "build"`), 0644)

		config := plan.NewProjectConfigurationFromFs(fs)
		assert.Equal(t, optional.Some("build"), plan.GetProjectConfigValue(config, "sm", "build_command"))
	})
}

func TestGetProjectConfigValue_None(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.toml", []byte(`[project]
`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs)
	assert.Equal(t, optional.None[string](), plan.GetProjectConfigValue(config, "", "build_command"))
}
