package php_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// TODO: coverage of GetPHPVersion
func TestGetPHPVersion_NoComposer(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")

	v := php.GetPHPVersion(config, fs)
	assert.Equal(t, v, php.DefaultPHPVersion)
}

func TestGetPHPVersion_NoVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")

	v := php.GetPHPVersion(config, fs)
	assert.Equal(t, v, php.DefaultPHPVersion)
}

func TestGetPHPVersion_EmptyVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"php": ""
		}
	}`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")

	v := php.GetPHPVersion(config, fs)
	assert.Equal(t, v, php.DefaultPHPVersion)
}

func TestDetermineProjectFramework_NoComposer(t *testing.T) {
	fs := afero.NewMemMapFs()

	framework := php.DetermineProjectFramework(fs)
	assert.Equal(t, framework, types.PHPFrameworkNone)
}

func TestDetermineProjectFramework_Laravel(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"laravel/framework": "^8.0"
		}
	}`), 0o644)

	framework := php.DetermineProjectFramework(fs)
	assert.Equal(t, framework, types.PHPFrameworkLaravel)
}

func TestDetermineProjectFramework_ThinkPHP(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"topthink/framework": "^6.0"
		}
	}`), 0o644)

	framework := php.DetermineProjectFramework(fs)
	assert.Equal(t, framework, types.PHPFrameworkThinkphp)
}

func TestDetermineProjectFramework_CodeIgniter(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"codeigniter4/framework": "^4.0"
		}
	}`), 0o644)

	framework := php.DetermineProjectFramework(fs)
	assert.Equal(t, framework, types.PHPFrameworkCodeigniter)
}

func TestDetermineProjectFramework_Unknown(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"akakkakakakakakakakakkakaka": "^11.45.14"
		},
	}`), 0o644)

	framework := php.DetermineProjectFramework(fs)
	assert.Equal(t, framework, types.PHPFrameworkNone)
}

func TestDetermineStartCommand_CustomInConfig(t *testing.T) {
	const expectedCommand = "php artisan serve; _startup"

	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	config.Set(plan.ConfigStartCommand, expectedCommand)

	actualCommand := php.DetermineStartCommand(config)

	assert.Contains(t, actualCommand, expectedCommand)
}

func TestDetermineBuildCommand_Default(t *testing.T) {
	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	command := php.DetermineBuildCommand(config)

	assert.Equal(t, "", command)
}

func TestDetermineBuildCommand_CustomInConfig(t *testing.T) {
	const expectedCommand = "php bin/build"

	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigBuildCommand, expectedCommand)

	actualCommand := php.DetermineBuildCommand(config)

	assert.Equal(t, expectedCommand, actualCommand)
}
