package php

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGetRequire_Exist(t *testing.T) {
	cJSON := composerJSONSchema{
		Require: &map[string]string{
			"php": ">=7.2",
		},
	}

	v, ok := cJSON.GetRequire("php")
	assert.True(t, ok)
	assert.Equal(t, v, ">=7.2")
}

func TestGetRequire_NotExist(t *testing.T) {
	cJSON := composerJSONSchema{
		Require: &map[string]string{
			"php": ">=7.2",
		},
	}

	_, ok := cJSON.GetRequire("php2")
	assert.False(t, ok)
}

func TestGetRequire_NoRequire(t *testing.T) {
	cJSON := composerJSONSchema{}

	_, ok := cJSON.GetRequire("php2")
	assert.False(t, ok)
}

func TestGetRequireDev_Exist(t *testing.T) {
	cJSON := composerJSONSchema{
		RequireDev: &map[string]string{
			"php": ">=7.2",
		},
	}

	v, ok := cJSON.GetRequireDev("php")
	assert.True(t, ok)
	assert.Equal(t, v, ">=7.2")
}

func TestGetRequireDev_NotExist(t *testing.T) {
	cJSON := composerJSONSchema{
		RequireDev: &map[string]string{
			"php": ">=7.2",
		},
	}

	_, ok := cJSON.GetRequireDev("php2")
	assert.False(t, ok)
}

func TestGetRequireDev_NoRequire(t *testing.T) {
	cJSON := composerJSONSchema{}

	_, ok := cJSON.GetRequireDev("php2")
	assert.False(t, ok)
}

func TestParseComposerJSON(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"php": ">=7.2"
		},
		"require-dev": {
			"phpunit/phpunit": "^8.5"
		}
	}`), 0o644)

	cJSON, err := parseComposerJSON(fs)
	assert.NoError(t, err)
	assert.Equal(t, cJSON.Name, "test")
	assert.Equal(t, cJSON.Require, &map[string]string{
		"php": ">=7.2",
	})
	assert.Equal(t, cJSON.RequireDev, &map[string]string{
		"phpunit/phpunit": "^8.5",
	})
}

func TestParseComposerJSON_NotExist(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := parseComposerJSON(fs)
	assert.Error(t, err)
}

func TestParseComposerJSON_WithoutRequireDev(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require": {
			"php": ">=7.2"
		}
	}`), 0o644)

	cJSON, err := parseComposerJSON(fs)
	assert.NoError(t, err)
	assert.Equal(t, cJSON.Name, "test")
	assert.Equal(t, cJSON.Require, &map[string]string{
		"php": ">=7.2",
	})
	assert.Nil(t, cJSON.RequireDev)
}

func TestParseComposerJSON_WithoutRequire(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"name": "test",
		"require-dev": {
			"phpunit/phpunit": "^8.5"
		}
	}`), 0o644)

	cJSON, err := parseComposerJSON(fs)
	assert.NoError(t, err)
	assert.Equal(t, cJSON.Name, "test")
	assert.Nil(t, cJSON.Require)
	assert.Equal(t, cJSON.RequireDev, &map[string]string{
		"phpunit/phpunit": "^8.5",
	})
}

func TestParseComposerJSON_WithoutName(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "composer.json", []byte(`{
		"require": {
			"php": ">=7.2"
		},
		"require-dev": {
			"phpunit/phpunit": "^8.5"
		}
	}`), 0o644)

	cJSON, err := parseComposerJSON(fs)
	assert.NoError(t, err)
	assert.Equal(t, cJSON.Name, "")
	assert.Equal(t, cJSON.Require, &map[string]string{
		"php": ">=7.2",
	})
	assert.Equal(t, cJSON.RequireDev, &map[string]string{
		"phpunit/phpunit": "^8.5",
	})
}
