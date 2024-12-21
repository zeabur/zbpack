package golang

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestGetEntry(t *testing.T) {
	t.Parallel()

	t.Run("without entry", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		entry := getEntry(ctx)
		assert.Equal(t, "", entry)
	})

	t.Run("with root main.go", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")
		_ = afero.WriteFile(fs, "main.go", nil, 0o644)

		ctx := &goPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "server",
		}

		entry := getEntry(ctx)
		assert.Equal(t, "", entry)
	})

	t.Run("with submodule", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")

		_ = afero.WriteFile(fs, "cmd/server/main.go", nil, 0o644)

		ctx := &goPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "server",
		}

		entry := getEntry(ctx)
		assert.Equal(t, "cmd/server/main.go", entry)
	})

	t.Run("with entry", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")
		config.Set(ConfigGoEntry, "cmd/server/main.go")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		entry := getEntry(ctx)
		assert.Equal(t, "cmd/server/main.go", entry)
	})
}

func TestGetBuildCommand(t *testing.T) {
	t.Parallel()

	t.Run("without build command", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		buildCommand := getBuildCommand(ctx)
		assert.Equal(t, "", buildCommand)
	})

	t.Run("with build command", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")
		config.Set(plan.ConfigBuildCommand, "go generate ./...")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		buildCommand := getBuildCommand(ctx)
		assert.Equal(t, "go generate ./...", buildCommand)
	})
}

func TestIsCgoEnabled(t *testing.T) {
	// Clear the user's environment variables.
	require.NoError(t, os.Setenv("CGO_ENABLED", ""))

	t.Run("default", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		assert.False(t, isCgoEnabled(ctx))
	})

	t.Run("with config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")
		config.Set(ConfigCgo, true)

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		assert.True(t, isCgoEnabled(ctx))
	})

	t.Run("with env", func(t *testing.T) {
		require.NoError(t, os.Setenv("CGO_ENABLED", "1"))

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "test")

		ctx := &goPlanContext{
			Src:    fs,
			Config: config,
		}

		assert.True(t, isCgoEnabled(ctx))
	})
}
