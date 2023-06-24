package php_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/pkg/types"
)

// TODO: coverage of GetPHPVersion
func TestGetPHPVersion_NoComposer(t *testing.T) {
	fs := afero.NewMemMapFs()

	v := php.GetPHPVersion(fs)
	assert.Equal(t, v, php.DefaultPHPVersion)
}

func TestGetPHPVersion_NoVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)

	v := php.GetPHPVersion(fs)
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

	v := php.GetPHPVersion(fs)
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

	app := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationDefault)
}

func TestDetermineApplication_Unknown(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)

	app := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationDefault)
}

func TestDetermineApplication_AcgFaka(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "lizhipay/acg-faka"
	}`), 0o644)

	app := php.DetermineApplication(fs)
	assert.Equal(t, app, types.PHPApplicationAcgFaka)
}
