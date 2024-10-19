package bun_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/bun"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestBunVersion(t *testing.T) {
	t.Parallel()

	t.Run("unspecified", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		ctx := bun.CreateBunContext(bun.GetMetaOptions{
			Src:    fs,
			Config: plan.NewProjectConfigurationFromFs(fs, ""),
		})

		version := bun.DetermineVersion(ctx)
		assert.Equal(t, "latest", version)
	})

	t.Run("exact version", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{"engines":{"bun":"1.2.3"}}`), 0o644)

		ctx := bun.CreateBunContext(bun.GetMetaOptions{
			Src:    fs,
			Config: plan.NewProjectConfigurationFromFs(fs, ""),
		})

		version := bun.DetermineVersion(ctx)
		assert.Equal(t, "1.2", version)
	})

	t.Run("range version", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{"engines":{"bun":"^1.2.3"}}`), 0o644)

		ctx := bun.CreateBunContext(bun.GetMetaOptions{
			Src:    fs,
			Config: plan.NewProjectConfigurationFromFs(fs, ""),
		})

		version := bun.DetermineVersion(ctx)
		assert.Equal(t, "1", version)
	})
}
