package python

import (
	"fmt"
	"strings"
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
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(strings.TrimSpace(`
[tool.poetry]
name = "poetry-demo"
version = "0.1.0"
description = ""
authors = ["Your Name <you@example.com>"]
readme = "README.md"
packages = [{include = "poetry_demo"}]

[tool.poetry.dependencies]
python = "^3.10"
flask = "^2.3.2"


[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"


`)), 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPoetry, pm)
}

func TestPackageManager_Pdm(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(strings.TrimSpace(`
[project]
name = ""
version = ""
description = ""
authors = [
    {name = "", email = ""},
]
dependencies = [
    "flask>=2.3.2",
]
requires-python = ">=3.8"
license = {text = "MIT"}

`)), 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPdm, pm)
}

func TestPackageManager_PoetryWithOldRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(strings.TrimSpace(`
[tool.poetry]
name = "poetry-demo"
version = "0.1.0"
description = ""
authors = ["Your Name <you@example.com>"]
readme = "README.md"
packages = [{include = "poetry_demo"}]

[tool.poetry.dependencies]
python = "^3.10"
flask = "^2.3.2"


[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"

`)), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPoetry, pm)
}

func TestPackageManager_PdmWithOldRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(strings.TrimSpace(`
[project]
name = ""
version = ""
description = ""
authors = [
    {name = "", email = ""},
]
dependencies = [
    "flask>=2.3.2",
]
requires-python = ">=3.8"
license = {text = "MIT"}

`)), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPdm, pm)
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

func TestPackageManager_PipenvWithOldRequirements_FixedOrder(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src: fs,
	}

	for i := 0; i < 1_000; i++ {
		pm := DeterminePackageManager(ctx)
		assert.Equal(t, types.PythonPackageManagerPipenv, pm, fmt.Sprintf("in the test round %d", i))
		ctx.PackageManager = nil
	}
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

func TestHasDependency_Pip_HasMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("mysqlclient==1.145.14"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Pip_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("mysqlalternative==19.19.810"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.False(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Pipenv_DirectlyUseMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte(strings.TrimSpace(`
[packages]
mysqlclient = "*"
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Pipenv_DependOnMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "Pipfile.lock", []byte(strings.TrimSpace(`
{
	"_meta": {
	"hash": {
		"sha256": "e34c3a87a1be2067ce73dbe50cae2e971a0190f15e361c32c82371256b2045b2"
	},
	"pipfile-spec": 6,
	"requires": {
		"python_version": "3.6"
	},
	"sources": [
		{
		"name": "pypi",
		"url": "https://pypi.python.org/simple",
		"verify_ssl": true
		}
	]
	},
	"default": {
	"mysqlclient": {
		"hashes": [
		"sha256:1d987a998c75633c40847cc966fcf5904906c920a7f17ef374f5aa4282abd304",
		"sha256:51fcb31174be6e6664c5f69e3e1691a2d72a1a12e90f872cbdb1567eb47b6519"
		],
		"version": "==12.34.56"
	}
	}
}
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Pipfile_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte(strings.TrimSpace(`
[packages]
mysqlalt = "*"
`)), 0o644)
	_ = afero.WriteFile(fs, "Pipfile.lock", []byte(strings.TrimSpace(`
{
	"_meta": {
	"hash": {
		"sha256": "e34c3a87a1be2067ce73dbe50cae2e971a0190f15e361c32c82371256b2045b2"
	},
	"pipfile-spec": 6,
	"requires": {
		"python_version": "3.6"
	},
	"sources": [
		{
		"name": "pypi",
		"url": "https://pypi.python.org/simple",
		"verify_ssl": true
		}
	]
	},
	"default": {
	"mysqlalt": {
		"hashes": [
		"sha256:1d987a998c75633c40847cc966fcf5904906c920a7f17ef374f5aa4282abd304",
		"sha256:51fcb31174be6e6664c5f69e3e1691a2d72a1a12e90f872cbdb1567eb47b6519"
		],
		"version": "==12.34.56"
	}
	}
}
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.False(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Poetry_DirectlyUseMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(strings.TrimSpace(`
[tool.poetry.dependencies]
mysqlclient = "^12.34.56"
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Poetry_DependOnMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "poetry.lock", []byte(strings.TrimSpace(`
[[package]]
name = "mysqlclient"
version = "22.2.0"
description = "Classes Without Boilerplate"
category = "main"
optional = false
python-versions = ">=3.6"
files = [
	{file = "attrs-22.2.0-py3-none-any.whl", hash = "sha256:29e95c7f6778868dbd49170f98f8818f78f3dc5e0e37c0b1f474e3561b240836"},
	{file = "attrs-22.2.0.tar.gz", hash = "sha256:c9227bfc2f01993c03f68db37d1d15c9690188323c067c641f1a35ca58185f99"},
]
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Poetry_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(""), 0o644)
	_ = afero.WriteFile(fs, "poetry.lock", []byte(strings.TrimSpace(`
[[package]]
name = "attrs"
version = "22.2.0"
description = "Classes Without Boilerplate"
category = "main"
optional = false
python-versions = ">=3.6"
files = [
	{file = "attrs-22.2.0-py3-none-any.whl", hash = "sha256:29e95c7f6778868dbd49170f98f8818f78f3dc5e0e37c0b1f474e3561b240836"},
	{file = "attrs-22.2.0.tar.gz", hash = "sha256:c9227bfc2f01993c03f68db37d1d15c9690188323c067c641f1a35ca58185f99"},
]
`)), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.False(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_CaseInsensitive(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("FOO"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}
