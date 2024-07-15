package rust

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestNeedOpenssl_CargoLockfile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Cargo.lock", []byte("openssl"), 0o644)

	assert.True(t, needOpenssl(fs))
}

func TestNeedOpenssl_CargoTomlfile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Cargo.toml", []byte("openssl"), 0o644)

	assert.True(t, needOpenssl(fs))
}

func TestNeedOpenssl_None(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Cargo.toml", []byte(""), 0o644)

	assert.False(t, needOpenssl(fs))
}

func TestGetEntry(t *testing.T) {
	t.Parallel()

	t.Run("no submodule name", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")

		entry := getEntry(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, "main", entry)
	})

	t.Run("with submodule name", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "submodule")

		entry := getEntry(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "submodule",
		})
		assert.Equal(t, "submodule", entry)
	})

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigRustEntry, "configured")

		entry := getEntry(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, "configured", entry)
	})
}

func TestGetAppDir(t *testing.T) {
	t.Parallel()

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigRustAppDir, "configured")

		appDir := getAppDir(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, "configured", appDir)
	})

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")

		appDir := getAppDir(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, ".", appDir)
	})

	t.Run("set as '/'", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigRustAppDir, "/")

		appDir := getAppDir(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, ".", appDir)
	})
}

func TestGetAssets(t *testing.T) {
	t.Parallel()

	t.Run("pass as array (json)", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigRustAssets, []string{"a", "b"})

		assets := getAssets(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, []string{"a", "b"}, assets)
	})

	t.Run("pass as string", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigRustAssets, "a b c")

		assets := getAssets(&rustPlanContext{
			Src:           fs,
			Config:        config,
			SubmoduleName: "",
		})
		assert.Equal(t, []string{"a", "b", "c"}, assets)
	})

	t.Run(".zeabur-preserve", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, ".zeabur-preserve", []byte("a\nb\nc"), 0o644)

		assets := getAssets(&rustPlanContext{
			Src:           fs,
			Config:        plan.NewProjectConfigurationFromFs(fs, ""),
			SubmoduleName: "",
		})
		assert.Equal(t, []string{"a", "b", "c"}, assets)
	})

	t.Run("no assets", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		assets := getAssets(&rustPlanContext{
			Src:           fs,
			Config:        plan.NewProjectConfigurationFromFs(fs, ""),
			SubmoduleName: "",
		})
		assert.Empty(t, assets)
	})
}
