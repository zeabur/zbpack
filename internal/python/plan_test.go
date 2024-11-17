package python

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run `go-snaps` will sort snapshots
	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestPackageManager_Pip(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPip, pm)
}

func TestPackageManager_Pipenv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
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
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
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
	_ = afero.WriteFile(fs, "pdm.lock", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPdm, pm)
}

func TestPackageManager_Rye(t *testing.T) {
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
	_ = afero.WriteFile(fs, "requirements.lock", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerRye, pm)
}

func TestPackageManager_Uv(t *testing.T) {
	t.Parallel()

	t.Run("contains uv.lock", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "pyproject.toml", nil, 0o644)
		_ = afero.WriteFile(fs, "uv.lock", nil, 0o644)

		ctx := &pythonPlanContext{
			Src:    fs,
			Config: plan.NewProjectConfigurationFromFs(fs, ""),
		}

		pm := DeterminePackageManager(ctx)

		assert.Equal(t, types.PythonPackageManagerUv, pm)
	})
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
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
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
	_ = afero.WriteFile(fs, "pdm.lock", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPdm, pm)
}

func TestPackageManager_PipenvWithOldRequirements(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}
	pm := DeterminePackageManager(ctx)

	assert.Equal(t, types.PythonPackageManagerPipenv, pm)
}

func TestPackageManager_PipenvWithOldRequirements_FixedOrder(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", nil, 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", nil, 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}

	for i := 0; i < 1_000; i++ {
		pm := DeterminePackageManager(ctx)
		assert.Equal(t, types.PythonPackageManagerPipenv, pm, fmt.Sprintf("in the test round %d", i))
		ctx.PackageManager = nil
	}
}

func TestPackageManager_Specified(t *testing.T) {
	t.Parallel()

	supportedPackageManagers := []types.PythonPackageManager{
		types.PythonPackageManagerPip,
		types.PythonPackageManagerPipenv,
		types.PythonPackageManagerPoetry,
		types.PythonPackageManagerPdm,
		types.PythonPackageManagerRye,
		types.PythonPackageManagerUv,
	}

	for _, pm := range supportedPackageManagers {
		pm := pm

		t.Run(string(pm), func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			config := plan.NewProjectConfigurationFromFs(fs, "")

			config.Set(ConfigPythonPackageManager, string(pm))

			ctx := pythonPlanContext{
				Src:    fs,
				Config: config,
			}

			determinedPm := DeterminePackageManager(&ctx)

			assert.Equal(t, pm, determinedPm)
		})
	}
}

func TestDetermineInstallCmd_Snapshot(t *testing.T) {
	const (
		WithWsgi              = "with-wsgi"
		WithFastapi           = "with-fastapi"
		WithTornado           = "with-tornado"
		WithStaticDjango      = "with-static-django"
		WithStaticNginx       = "with-static-nginx"
		WithStaticNginxDjango = "with-static-nginx-django"
		WithStreamlitEntry    = "with-streamlit-entry"
		None                  = "none"
	)

	for _, pm := range []types.PythonPackageManager{
		types.PythonPackageManagerPipenv,
		types.PythonPackageManagerPoetry,
		types.PythonPackageManagerPip,
		types.PythonPackageManagerPdm,
		types.PythonPackageManagerUnknown,
	} {
		pm := pm
		for _, mode := range []string{
			WithWsgi,
			WithFastapi,
			WithTornado,
			WithStaticNginx,
			WithStaticDjango,
			WithStaticNginxDjango,
			WithStreamlitEntry,
			None,
		} {
			mode := mode
			t.Run(string(pm)+"-"+mode, func(t *testing.T) {
				t.Parallel()

				fs := afero.NewMemMapFs()
				config := plan.NewProjectConfigurationFromFs(fs, "")

				ctx := pythonPlanContext{
					Src:            fs,
					Config:         config,
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

				if mode == WithTornado {
					ctx.Framework = optional.Some(types.PythonFrameworkTornado)
				}

				if mode == WithStaticNginx {
					ctx.Wsgi = optional.Some("wsgi.py")
					ctx.Static = optional.Some(StaticInfo{
						Flag:          StaticModeNginx,
						StaticURLPath: "/static",
						StaticHostDir: "/app/static",
					})
				}

				if mode == WithStaticDjango {
					ctx.Wsgi = optional.Some("wsgi.py")
					ctx.Framework = optional.Some(types.PythonFrameworkDjango)
					ctx.Static = optional.Some(StaticInfo{
						Flag:          StaticModeDjango,
						StaticURLPath: "/static",
						StaticHostDir: "/app/static",
					})
				}

				if mode == WithStaticNginxDjango {
					ctx.Wsgi = optional.Some("wsgi.py")
					ctx.Framework = optional.Some(types.PythonFrameworkDjango)
					ctx.Static = optional.Some(StaticInfo{
						Flag:          StaticModeNginx | StaticModeDjango,
						StaticURLPath: "/static",
						StaticHostDir: "/app/static",
					})
				}

				if mode == WithStreamlitEntry {
					ctx.StreamlitEntry = optional.Some("streamlit_app.py")
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
		WithStatic  = "with-static"
		None        = "none"
	)

	for _, pm := range []types.PythonPackageManager{
		types.PythonPackageManagerPipenv,
		types.PythonPackageManagerPoetry,
		types.PythonPackageManagerPip,
		types.PythonPackageManagerUnknown,
	} {
		pm := pm
		for _, mode := range []string{WithWsgi, WithFastapi, WithStatic, None} {
			mode := mode
			t.Run(string(pm)+"-"+mode, func(t *testing.T) {
				t.Parallel()

				fs := afero.NewMemMapFs()

				ctx := pythonPlanContext{
					Src:            fs,
					Config:         plan.NewProjectConfigurationFromFs(fs, ""),
					PackageManager: optional.Some(pm),
					Entry:          optional.Some("app.py"),
				}

				if mode == WithWsgi || mode == WithFastapi {
					ctx.Wsgi = optional.Some("wsgi.py")
				} else {
					ctx.Wsgi = optional.Some("") // fake cache
				}

				if mode == WithStatic {
					ctx.Wsgi = optional.Some("wsgi.py")
					ctx.Static = optional.Some(StaticInfo{
						Flag:          StaticModeNginx,
						StaticURLPath: "/static",
						StaticHostDir: "/app/static",
					})
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_Pip_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("mysqlalternative==19.19.810"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
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
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.False(t, HasDependency(ctx, "mysqlclient"))
}

func TestHasDependency_CaseInsensitive(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("FOO"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasDependency(ctx, "foo"))
	assert.False(t, HasDependency(ctx, "bar"))
}

func TestHasDependency_Uv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(`
[project]
name = "midexam"
version = "0.1.0"
description = "Add your description here"
readme = "README.md"
requires-python = ">=3.12"
dependencies = [
    "matplotlib>=3.9.2",
    "notebook>=7.2.2",
    "pandas[output-formatting,performance,plot]>=2.2.3",
    "polars[numpy,pandas,plot]>=1.13.1",
    "scikit-learn>=1.5.2",
    "seaborn>=0.13.2",
]`), 0o644)
	_ = afero.WriteFile(fs, "uv.lock", []byte(`version = 1
requires-python = ">=3.12"

[[package]]
name = "altair"
version = "5.4.1"
source = { registry = "https://pypi.org/simple" }
dependencies = [
    { name = "jinja2" },
    { name = "jsonschema" },
    { name = "narwhals" },
    { name = "packaging" },
    { name = "typing-extensions", marker = "python_full_version < '3.13'" },
]
sdist = { url = "https://files.pythonhosted.org/packages/ae/09/38904138a49f29e529b61b4f39954a6837f443d828c1bc57814be7bd4813/altair-5.4.1.tar.gz", hash = "sha256:0ce8c2e66546cb327e5f2d7572ec0e7c6feece816203215613962f0ec1d76a82", size = 636465 }
wheels = [
    { url = "https://files.pythonhosted.org/packages/9b/52/4a86a4fa1cc2aae79137cc9510b7080c3e5aede2310d14fae5486feec7f7/altair-5.4.1-py3-none-any.whl", hash = "sha256:0fb130b8297a569d08991fb6fe763582e7569f8a04643bbd9212436e3be04aef", size = 658150 },
]`), 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}

	assert.True(t, HasDependency(ctx, "matplotlib"))
	assert.True(t, HasDependency(ctx, "notebook"))
	assert.True(t, HasDependency(ctx, "altair"))
}

func TestHasExplicitDependency_Pip(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "requirements.txt", []byte("foo"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPip),
	}

	assert.True(t, HasExplicitDependency(ctx, "foo"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestHasExplicitDependency_Pipenv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Pipfile", []byte(`
[[source]]
url = "https://pypi.python.org/simple"
verify_ssl = true
name = "pypi"

[packages]
requests = "*"


[dev-packages]
pytest = "*"`), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPipenv),
	}

	assert.True(t, HasExplicitDependency(ctx, "requests"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestHasExplicitDependency_Poetry(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(`
[tool.poetry]
name = "changchun api"
version = "0.1.0"
description = ""
authors = ["Your Name <you@example.com>"]
readme = "README.md"

[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.95.2"
psycopg2 = "^2.9.6"


[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"`), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPoetry),
	}

	assert.True(t, HasExplicitDependency(ctx, "fastapi"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestHasExplicitDependency_Pdm(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(`

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
license = {text = "MIT"}`), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerPdm),
	}

	assert.True(t, HasExplicitDependency(ctx, "flask"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestHasExplicitDependency_Rye(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(`

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
license = {text = "MIT"}`), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerRye),
	}

	assert.True(t, HasExplicitDependency(ctx, "flask"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestHasExplicitDependency_Uv(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "pyproject.toml", []byte(`
[project]
name = "midexam"
version = "0.1.0"
description = "Add your description here"
readme = "README.md"
requires-python = ">=3.12"
dependencies = [
    "matplotlib>=3.9.2",
    "notebook>=7.2.2",
    "pandas[output-formatting,performance,plot]>=2.2.3",
    "polars[numpy,pandas,plot]>=1.13.1",
    "scikit-learn>=1.5.2",
    "seaborn>=0.13.2",
]`), 0o644)
	_ = afero.WriteFile(fs, "uv.lock", []byte(`version = 1
requires-python = ">=3.12"

[[package]]
name = "altair"
version = "5.4.1"
source = { registry = "https://pypi.org/simple" }
dependencies = [
    { name = "jinja2" },
    { name = "jsonschema" },
    { name = "narwhals" },
    { name = "packaging" },
    { name = "typing-extensions", marker = "python_full_version < '3.13'" },
]
sdist = { url = "https://files.pythonhosted.org/packages/ae/09/38904138a49f29e529b61b4f39954a6837f443d828c1bc57814be7bd4813/altair-5.4.1.tar.gz", hash = "sha256:0ce8c2e66546cb327e5f2d7572ec0e7c6feece816203215613962f0ec1d76a82", size = 636465 }
wheels = [
    { url = "https://files.pythonhosted.org/packages/9b/52/4a86a4fa1cc2aae79137cc9510b7080c3e5aede2310d14fae5486feec7f7/altair-5.4.1-py3-none-any.whl", hash = "sha256:0fb130b8297a569d08991fb6fe763582e7569f8a04643bbd9212436e3be04aef", size = 658150 },
]`), 0o644)

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: plan.NewProjectConfigurationFromFs(fs, ""),
	}

	assert.True(t, HasExplicitDependency(ctx, "matplotlib"))
	assert.True(t, HasExplicitDependency(ctx, "notebook"))
	assert.False(t, HasExplicitDependency(ctx, "altair"))
}

func TestHasExplicitDependency_Unknown(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.False(t, HasExplicitDependency(ctx, "flask"))
	assert.False(t, HasExplicitDependency(ctx, "bar"))
}

func TestDetermineStreamlitEntry_ByFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "streamlit_app.py", []byte(`import streamlit as st
x = st.slider("Select a value")
st.write(x, "squared is", x * x)`), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("streamlit"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Equal(t, "streamlit_app.py", determineStreamlitEntry(ctx))
}

func TestDetermineStreamlitEntry_ByConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zeabur_streamlit_demo.py", []byte(`import streamlit as st
x = st.slider("Select a value")
st.write(x, "squared is", x * x)`), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("streamlit"), 0o644)
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{"streamlit": {"entry": "zeabur_streamlit_demo.py"}}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Equal(t, "zeabur_streamlit_demo.py", determineStreamlitEntry(ctx))
}

func TestDetermineStreamlitEntry_ConfigPrecedeConvention(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zeabur_streamlit_demo.py", []byte(`import streamlit as st
x = st.slider("Select a value")
st.write(x, "squared is", x * x)`), 0o644)
	_ = afero.WriteFile(fs, "app.py", []byte(`print('not me')`), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("streamlit"), 0o644)
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{"streamlit": {"entry": "zeabur_streamlit_demo.py"}}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Equal(t, "zeabur_streamlit_demo.py", determineStreamlitEntry(ctx))
}

func TestDetermineStreamlitEntry_Cache(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zeabur_streamlit_demo.py", []byte(`import streamlit as st
x = st.slider("Select a value")
st.write(x, "squared is", x * x)`), 0o644)
	_ = afero.WriteFile(fs, "app.py", []byte(`print('not me')`), 0o644)
	_ = afero.WriteFile(fs, "requirements.txt", []byte("streamlit"), 0o644)
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{"streamlit": {"entry": "zeabur_streamlit_demo.py"}}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Equal(t, "zeabur_streamlit_demo.py", determineStreamlitEntry(ctx))
	assert.Equal(t, "zeabur_streamlit_demo.py", ctx.StreamlitEntry.Unwrap())
}

func TestDetermineWsgi(t *testing.T) {
	t.Parallel()

	matchedCases := []string{
		"app = FastAPI()",
		"    app = FastAPI()",
		"app = FastAPI(\n\t# test\n)",
		"app=FastAPI(\n\tname='app'\n)",
	}

	notMatchedCases := []string{
		"app = FastAPI",
		"# FastAPI test",
		"# app = FastAPI test",
		"app=FastAPI",
	}

	for _, c := range matchedCases {
		c := c

		t.Run("matched-"+c, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, "main.py", []byte(c), 0o644)

			ctx := &pythonPlanContext{
				Src:            fs,
				Config:         plan.NewProjectConfigurationFromFs(fs, ""),
				PackageManager: optional.Some(types.PythonPackageManagerUnknown),
				Framework:      optional.Some(types.PythonFrameworkFastapi),
			}

			assert.Equal(t, "main:app", DetermineWsgi(ctx))
		})
	}

	for _, c := range notMatchedCases {
		c := c
		t.Run("not-matched-"+c, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, "main.py", []byte(c), 0o644)

			ctx := &pythonPlanContext{
				Src:            fs,
				Config:         plan.NewProjectConfigurationFromFs(fs, ""),
				PackageManager: optional.Some(types.PythonPackageManagerUnknown),
				Framework:      optional.Some(types.PythonFrameworkFastapi),
			}

			assert.Equal(t, "", DetermineWsgi(ctx))
		})
	}
}

func TestDeterminePythonVersion_Pipenv(t *testing.T) {
	t.Parallel()

	pipFile := []struct {
		testname string
		content  string
		expect   string
	}{
		{
			testname: "python_version with spaces",
			content: strings.TrimSpace(`
[[source]]
url = "https://pypi.python.org/simple"
verify_ssl = true
name = "pypi"

[requires]
python_version = "3.8"
`),
			expect: "3.8",
		},
		{
			testname: "python_version without spaces",
			content: strings.TrimSpace(`
[[source]]
url = "https://pypi.python.org/simple"
verify_ssl = true
name = "pypi"

[requires]
python_version="3.8"
`),
			expect: "3.8",
		},
		{
			testname: "python_version with two digit minor version",
			content: strings.TrimSpace(`
[[source]]
url = "https://pypi.python.org/simple"
verify_ssl = true
name = "pypi"

[requires]
python_version = "3.12"
`),
			expect: "3.12",
		},
	}

	for _, p := range pipFile {
		p := p
		t.Run(p.testname, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, "Pipfile", []byte(p.content), 0o644)

			ctx := &pythonPlanContext{
				Src:    fs,
				Config: plan.NewProjectConfigurationFromFs(fs, ""),
			}

			assert.Equal(t, p.expect, determinePythonVersion(ctx))
		})
	}
}

func TestDeterminePythonVersion_Customized(t *testing.T) {
	fs := afero.NewMemMapFs()
	conf := plan.NewProjectConfigurationFromFs(fs, "")
	conf.Set(ConfigPythonVersion, "3.12345.1")

	ctx := &pythonPlanContext{
		Src:    fs,
		Config: conf,
	}

	assert.Equal(t, "3.12345", determinePythonVersion(ctx))
}

func TestDetermineAptDependencies_Nodejs(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte("{}"), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Contains(t, determineAptDependencies(ctx), "nodejs")
	assert.Contains(t, determineAptDependencies(ctx), "npm")
}

func TestDetermineBuildCommand_NPMBuild(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{"scripts": {"build": "echo 'hi'"}}`), 0o644)

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         plan.NewProjectConfigurationFromFs(fs, ""),
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Contains(t, determineBuildCmd(ctx), "npm install && npm run build")
}

func TestDetermineBuildCommand_Custom(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{"scripts": {"build": "echo 'hi'"}}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigBuildCommand, "echo 'hello'")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Contains(t, determineBuildCmd(ctx), "echo 'hello'")
}

func TestDetermineStartCommand_Custom(t *testing.T) {
	fs := afero.NewMemMapFs()

	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set(plan.ConfigStartCommand, "echo 'hello'")

	ctx := &pythonPlanContext{
		Src:            fs,
		Config:         config,
		PackageManager: optional.Some(types.PythonPackageManagerUnknown),
	}

	assert.Contains(t, determineStartCmd(ctx), "echo 'hello'")
	assert.Contains(t, determineStartCmd(ctx), "_startup()") // should have the default startup function
}
