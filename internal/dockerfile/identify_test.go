package dockerfile_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/dockerfile"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestMatchDockerfile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	config := plan.NewProjectConfigurationFromFs(fs, "")

	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	identifier := dockerfile.NewIdentifier()
	assert.True(t, identifier.Match(plan.MatchContext{
		Source: fs,
		Config: config,
	}))
}

func TestMatchDockerfileWithSubmoduleName(t *testing.T) {
	t.Parallel()

	t.Run("suffix", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")

		_ = afero.WriteFile(fs, "Dockerfile.Subm", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})

	t.Run("prefix", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")

		_ = afero.WriteFile(fs, "Subm.Dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})

	t.Run("case insensitive", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")

		_ = afero.WriteFile(fs, "dockerfile.subm", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})
}

func TestMatchDockerfileWithCustomName(t *testing.T) {
	t.Parallel()

	t.Run("suffix", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.name", "custom")

		_ = afero.WriteFile(fs, "Dockerfile.custom", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})

	t.Run("prefix", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.name", "custom")

		_ = afero.WriteFile(fs, "custom.Dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})

	t.Run("case insensitive", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.name", "custom")

		_ = afero.WriteFile(fs, "custom.dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})

	t.Run("not existing", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.name", "custom")

		identifier := dockerfile.NewIdentifier()
		assert.False(t, identifier.Match(plan.MatchContext{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		}))
	})
}

func TestMatchDockerfileWithCustomPath(t *testing.T) {
	t.Parallel()

	t.Run("root directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.path", "custom.Dockerfile")

		_ = afero.WriteFile(fs, "custom.Dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source: fs,
			Config: config,
		}))
	})

	t.Run("with root prefix", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.path", "/custom.Dockerfile")

		_ = afero.WriteFile(fs, "custom.Dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source: fs,
			Config: config,
		}))
	})

	t.Run("submodule directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.path", "docker/custom.Dockerfile")

		_ = afero.WriteFile(fs, "docker/custom.Dockerfile", []byte("FROM alpine"), 0o644)

		identifier := dockerfile.NewIdentifier()
		assert.True(t, identifier.Match(plan.MatchContext{
			Source: fs,
			Config: config,
		}))
	})

	t.Run("not existing", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set("dockerfile.path", "custom.Dockerfile")

		identifier := dockerfile.NewIdentifier()
		assert.False(t, identifier.Match(plan.MatchContext{
			Source: fs,
			Config: config,
		}))
	})
}
