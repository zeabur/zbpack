package php

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDetermineAptDependencies_None(t *testing.T) {
	fs := afero.NewMemMapFs()

	deps := DetermineAptDependencies(fs)
	assert.Empty(t, deps)
}

func TestDetermineAptDependencies_NoRequire(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Empty(t, deps)
}

func TestDetermineAptDependencies_EmptyRequire(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Empty(t, deps)
}

func TestDetermineAptDependencies_RequireOpenssl(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"ext-openssl": "*"
		}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, depMap["ext-openssl"], deps)
}

func TestDetermineAptDependencies_RequireZip(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"ext-zip": "*"
		}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, depMap["ext-zip"], deps)
}

func TestDetermineAptDependencies_RequireCurl(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"ext-curl": "*"
		}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, depMap["ext-curl"], deps)
}

func TestDetermineAptDependencies_RequireGd(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"ext-gd": "*"
		}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, depMap["ext-gd"], deps)
}

func TestDetermineAptDependencies_Unknown(t *testing.T) {
	fs := afero.NewMemMapFs()

	deps := DetermineAptDependencies(fs)
	assert.Empty(t, deps)
}
