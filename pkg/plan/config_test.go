package plan_test

import (
	"testing"

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
