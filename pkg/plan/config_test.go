package plan_test

import (
	"testing"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestProjectConfiguration_Empty(t *testing.T) {
	t.Parallel()

	config := &plan.ViperProjectConfiguration{}
	assert.Equal(t, optional.None[any](), config.Get("laravel.test"))
}

func TestProjectConfiguration_ZbpackJsonExisted(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{"laravel":{"test": "owo" }}`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	assert.Equal(t, optional.Some[any]("owo"), config.Get("laravel.test"))
	assert.True(t, config.Get("laravel").IsSome())
	assert.True(t, config.Get("laravel.test").IsSome())
	assert.True(t, config.Get("laravel.owo").IsNone())
}

func TestProjectConfiguration_ZbpackJsonNotExisted(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()

	config := plan.NewProjectConfigurationFromFs(fs, "")

	assert.True(t, config.Get("laravel").IsNone())
	assert.True(t, config.Get("laravel.test").IsNone())
	assert.True(t, config.Get("laravel.owo").IsNone())
}

func TestProjectConfiguration_RootMalformed(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`I'm not JSON'`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	assert.True(t, config.Get("laravel").IsNone())
}

func TestProjectConfiguration_SubmoduleMalformed(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.hi.json", []byte(`I'm not JSON'`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "hi")

	assert.True(t, config.Get("laravel").IsNone())
}

func TestProjectConfiguration_SubmoduleMalformedWhileRootWorks(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{"hi": 1234}`), 0644)
	_ = afero.WriteFile(fs, "zbpack.hi.json", []byte(`I'm not JSON'`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "hi")

	assert.Equal(t, optional.Some[any](float64(1234)), config.Get("hi"))
}

func TestGet_Global(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()

	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{ "build_command": "build" }`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "global_test")
	assert.Equal(t, optional.Some[any]("build"), config.Get("build_command"))
}

func TestGet_Submodule(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.test.json", []byte(`{ "build_command": "build#test" }`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "test")
	assert.Equal(t, optional.Some[any]("build#test"), config.Get("build_command"))
}

func TestGet_SubmoduleOverride(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{ "build_command": "build" }`), 0644)
	_ = afero.WriteFile(fs, "zbpack.test.json", []byte(`{ "build_command": "build#test" }`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "test")
	assert.Equal(t, optional.Some[any]("build#test"), config.Get("build_command"))
}

func TestGet_SubmoduleFallback(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{ "build_command": "build" }`), 0644)
	_ = afero.WriteFile(fs, "zbpack.test.json", []byte(`{}`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "test")
	assert.Equal(t, optional.Some[any]("build"), config.Get("build_command"))
}

func TestGet_None(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{}`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "")
	assert.True(t, config.Get("build_command").IsNone())
}

func TestCastOptionValueOrNone(t *testing.T) {
	t.Parallel()

	assert.Equal(t, optional.Some[string]("owo"), plan.Cast(optional.Some[any]("owo"), cast.ToStringE))
	assert.Equal(t, optional.Some[int](1234), plan.Cast(optional.Some[any](1234.5), cast.ToIntE))
	assert.Equal(t, optional.None[uint](), plan.Cast(optional.Some[any](":)"), cast.ToUintE))
	assert.Equal(t, optional.None[string](), plan.Cast(optional.None[any](), cast.ToStringE))
}

func TestSet(t *testing.T) {
	t.Parallel()

	config := &plan.ViperProjectConfiguration{}
	config.Set("owo", "uwu")

	assert.Equal(t, optional.Some[any]("uwu"), config.Get("owo"))
}
