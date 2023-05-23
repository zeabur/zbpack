package python

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestPackageManager_Pip(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPip, pm)
}

func TestPackageManager_Pipenv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPipenv, pm)
}

func TestPackageManager_Poetry(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPoetry, pm)
}

func TestPackageManager_PoetryWithOldRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPoetry, pm)
}

func TestPackageManager_PipenvWithOldRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPipenv, pm)
}

func TestDetermineInstallCmd_Snapshot(t *testing.T) {
	const (
		WithWsgi    = "with-wsgi"
		WithFastapi = "with-fastapi"
		None        = "none"
	)

	for _, pm := range []types.PackageManager{
		types.PythonPackageManagerPipenv,
		types.PythonPackageManagerPoetry,
		types.PythonPackageManagerPip,
		types.PythonPackageManagerUnknown,
	} {
		pm := pm
		for _, mode := range []string{WithWsgi, WithFastapi, None} {
			mode := mode
			t.Run(string(pm)+"-"+mode, func(t *testing.T) {
				t.Parallel()

				ctx := pythonPlanContext{
					PackageManager: optional.Some(pm),
				}

				if mode == WithWsgi || mode == WithFastapi {
					ctx.Wsgi = optional.Some("wsgi.py")
				} else {
					ctx.Wsgi = optional.Some("") // fake cache
				}

				if mode == WithFastapi {
					ctx.Framework = optional.Some(types.PythonFrameworkFastapi)
				} else {
					ctx.Framework = optional.Some(types.PythonFrameworkNone)
				}

				ic := determineInstallCmd(&ctx)
				snaps.MatchSnapshot(t, ic)
			})
		}
	}
}

func TestDetermineStartCmd_Snapshot(t *testing.T) {
	const (
		WithWsgi    = "with-wsgi"
		WithFastapi = "with-fastapi"
		None        = "none"
	)

	for _, pm := range []types.PackageManager{
		types.PythonPackageManagerPipenv,
		types.PythonPackageManagerPoetry,
		types.PythonPackageManagerPip,
		types.PythonPackageManagerUnknown,
	} {
		pm := pm
		for _, mode := range []string{WithWsgi, WithFastapi, None} {
			mode := mode
			t.Run(string(pm)+"-"+mode, func(t *testing.T) {
				t.Parallel()

				ctx := pythonPlanContext{
					PackageManager: optional.Some(pm),
					Entry:          optional.Some("app.py"),
				}

				if mode == WithWsgi || mode == WithFastapi {
					ctx.Wsgi = optional.Some("wsgi.py")
				} else {
					ctx.Wsgi = optional.Some("") // fake cache
				}

				if mode == WithFastapi {
					ctx.Framework = optional.Some(types.PythonFrameworkFastapi)
				} else {
					ctx.Framework = optional.Some(types.PythonFrameworkNone)
				}

				ic := determineStartCmd(&ctx)
				snaps.MatchSnapshot(t, ic)
			})
		}
	}
}

func TestHasDependency_Unknown(t *testing.T) {
	fs := afero.NewMemMapFs()

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	// should always False
	assert.False(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_Pip(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_Poetry(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_PoetryDep(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "poetry.lock", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_Pipenv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_PipenvDep(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile.lock", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_Pipenv_WithObsoleteRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte("foo"), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("bar"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_PipenvDep_WithObsoleteRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile.lock", []byte("foo"), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("bar"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

