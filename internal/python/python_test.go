package python

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsMysqlNeeded_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()

	assert.False(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Requirement_HasMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/requirements.txt", []byte("mysqlclient==1.145.14"), 0o644)

	assert.True(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Requirement_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/requirements.txt", []byte("mysqlalternative==19.19.810"), 0o644)

	assert.False(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Pipfile_DirectlyUseMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/Pipfile", []byte(strings.TrimSpace(`
[packages]
mysqlclient = "*"
`)), 0o644)

	assert.True(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Pipfile_DependOnMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/Pipfile", []byte(""), 0o644)
	afero.WriteFile(fs, "/src/Pipfile.lock", []byte(strings.TrimSpace(`
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

	assert.True(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Pipfile_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/Pipfile", []byte(strings.TrimSpace(`
[packages]
mysqlalt = "*"
`)), 0o644)
	afero.WriteFile(fs, "/src/Pipfile.lock", []byte(strings.TrimSpace(`
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

	assert.False(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Poetry_DirectlyUseMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/pyproject.toml", []byte(strings.TrimSpace(`
[tool.poetry.dependencies]
mysqlclient = "^12.34.56"
`)), 0o644)

	assert.True(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Poetry_DependOnMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/pyproject.toml", []byte(""), 0o644)
	afero.WriteFile(fs, "/src/poetry.lock", []byte(strings.TrimSpace(`
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

	assert.True(t, IsMysqlNeeded(fs))
}

func TestIsMysqlNeeded_Poetry_NoMysqlClient(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/src/pyproject.toml", []byte(""), 0o644)
	afero.WriteFile(fs, "/src/poetry.lock", []byte(strings.TrimSpace(`
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

	assert.False(t, IsMysqlNeeded(fs))
}
