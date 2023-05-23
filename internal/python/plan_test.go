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

