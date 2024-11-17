package python

import (
	"strings"

	"github.com/zeabur/zbpack/pkg/types"
)

func getPmInitCmd(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pip install pipenv"
	case types.PythonPackageManagerPoetry:
		return "pip install poetry"
	case types.PythonPackageManagerPdm:
		return "pip install pdm"
	case types.PythonPackageManagerRye: // managed with pip
	case types.PythonPackageManagerUv:
		return "pip install uv"
	}

	return ""
}

func getPmAddCmd(pm types.PythonPackageManager, deps ...string) string {
	if len(deps) == 0 {
		return ""
	}

	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pipenv install " + strings.Join(deps, " ")
	case types.PythonPackageManagerPoetry:
		return "poetry add " + strings.Join(deps, " ")
	case types.PythonPackageManagerPdm:
		return "pdm add " + strings.Join(deps, " ")
	case types.PythonPackageManagerRye: // managed with pip
	case types.PythonPackageManagerUv:
		return "uv add " + strings.Join(deps, " ")
	}

	return "pip install " + strings.Join(deps, " ")
}

func getPmInstallCmd(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPip:
		return "sed '/-e/d' requirements.txt | pip install -r /dev/stdin"
	case types.PythonPackageManagerPipenv:
		return "pipenv install"
	case types.PythonPackageManagerPoetry:
		return "poetry install"
	case types.PythonPackageManagerPdm:
		return "pdm install"
	case types.PythonPackageManagerRye:
		return "sed '/-e/d' requirements.lock | pip install -r /dev/stdin"
	case types.PythonPackageManagerUv:
		return "uv sync"
	}

	return ""
}

func getPmPostInstallCmd(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPip:
		// bind project directory in site_packages
		return "pip install -r requirements.txt"
	case types.PythonPackageManagerRye:
		return "pip install -r requirements.lock"
	}

	return ""
}

func getPmStartCmdPrefix(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pipenv run"
	case types.PythonPackageManagerPoetry:
		return "poetry run"
	case types.PythonPackageManagerPdm:
		return "pdm run"
	case types.PythonPackageManagerRye:
		return "" // unneeded
	case types.PythonPackageManagerUv:
		return "uv run"
	}

	return ""
}

func getPmDeclarationFile(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPip:
		return "requirements.txt"
	case types.PythonPackageManagerPipenv:
		return "Pipfile"
	case types.PythonPackageManagerPoetry, types.PythonPackageManagerPdm, types.PythonPackageManagerRye, types.PythonPackageManagerUv:
		return "pyproject.toml"
	}

	return ""
}

func getPmLockFile(pm types.PythonPackageManager) []string {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return []string{"Pipfile.lock"}
	case types.PythonPackageManagerPoetry:
		return []string{"poetry.lock"}
	case types.PythonPackageManagerPdm:
		return []string{"pdm.lock"}
	case types.PythonPackageManagerRye:
		return []string{"requirements.lock", ".python-version"}
	case types.PythonPackageManagerUv:
		return []string{"uv.lock", ".python-version"}
	}

	return nil
}
