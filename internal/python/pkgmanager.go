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
	}

	return ""
}

func getPmPostInstallCmd(pm types.PythonPackageManager) string {
	switch pm {
	case types.PythonPackageManagerPip:
		// bind project directory in site_packages
		return "pip install -r requirements.txt"
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
	}

	return ""
}

func getPmDeclarationFile(pm types.PythonPackageManager) []string {
	switch pm {
	case types.PythonPackageManagerPip:
		return []string{"requirements.txt", "pyproject.toml" /* newer projects */}
	case types.PythonPackageManagerPipenv:
		return []string{"Pipfile"}
	case types.PythonPackageManagerPoetry, types.PythonPackageManagerPdm:
		return []string{"pyproject.toml"}
	}

	return nil
}

func getPmLockFile(pm types.PythonPackageManager) []string {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return []string{"Pipfile.lock"}
	case types.PythonPackageManagerPoetry:
		return []string{"poetry.lock"}
	case types.PythonPackageManagerPdm:
		return []string{"pdm.lock"}
	}

	return nil
}
