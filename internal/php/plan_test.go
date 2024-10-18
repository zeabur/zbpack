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

func TestDetermineApplication_NoComposer(t *testing.T) {
	fs := afero.NewMemMapFs()

	app, kind := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationDefault)
	assert.Equal(t, types.PHPPropertyNone, kind&types.PHPPropertyComposer)
}

func TestDetermineApplication_UnknownWithComposer(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)

	app, kind := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationDefault)
	assert.NotEqual(t, types.PHPPropertyNone, kind&types.PHPPropertyComposer)
}

func TestDetermineApplication_AcgFaka(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "lizhipay/acg-faka"
	}`), 0o644)

	app, kind := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationAcgFaka)
	assert.NotEqual(t, types.PHPPropertyNone, kind&types.PHPPropertyComposer)
}

func TestDetermineStartCommand_Default(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	command := php.DetermineStartCommand(config)

	assert.Contains(t, command, "nginx; php-fpm")
}

func TestDetermineStartCommand_Swoole(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	config.Set(php.ConfigLaravelOctaneServer, php.OctaneServerSwoole)
	command := php.DetermineStartCommand(config)

	assert.Contains(t, command, "php artisan octane:start --server=swoole --host=0.0.0.0 --port=8080")
}

func TestDetermineStartCommand_Roadrunner(t *testing.T) {
	// unimplemented, so it should fall back to the default command

	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	config.Set(php.ConfigLaravelOctaneServer, php.OctaneServerRoadrunner)
	command := php.DetermineStartCommand(config)

	assert.Contains(t, command, "nginx; php-fpm")
}

func TestDetermineStartCommand_UnknownOctane(t *testing.T) {
	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
	config.Set(php.ConfigLaravelOctaneServer, "unknown")
	command := php.DetermineStartCommand(config)

	assert.Contains(t, command, "nginx; php-fpm")
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
	command := php.DetermineBuildCommand(fs, config)

	assert.Equal(t, "", command)
}

func TestDetermineBuildCommand_CustomInConfig(t *testing.T) {
	const expectedCommand = "php bin/build"

	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigBuildCommand, expectedCommand)

	actualCommand := php.DetermineBuildCommand(fs, config)

	assert.Equal(t, expectedCommand, actualCommand)
}

func TestDetermineBuildCommand_NPMBuild(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{
		"scripts": {
			"build": "npm run build"
		}
	}`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")
	command := php.DetermineBuildCommand(fs, config)

	assert.Equal(t, "npm install && npm run build", command)
}
