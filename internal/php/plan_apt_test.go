package php

// due to some internal logics, we need to do blackbox test
import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDetermineAptDependencies_None(t *testing.T) {
	fs := afero.NewMemMapFs()

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, deps, baseDep)
}

func TestDetermineAptDependencies_NoRequire(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test"
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, deps, baseDep)
}

func TestDetermineAptDependencies_EmptyRequire(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {}
	}`), 0o644)

	deps := DetermineAptDependencies(fs)
	assert.Equal(t, deps, baseDep)
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
	assert.Equal(t, deps, append(baseDep, depMap["ext-openssl"]...))
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
	assert.Equal(t, deps, append(baseDep, depMap["ext-zip"]...))
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
	assert.Equal(t, deps, append(baseDep, depMap["ext-curl"]...))
}
