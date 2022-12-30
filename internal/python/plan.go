package python

import (
	"os"
	"path"
	"strings"

	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineFramework(absPath string) PythonFramework {
	requirementsTxt, err := os.ReadFile(path.Join(absPath, "requirements.txt"))
	if err != nil {
		return PythonFrameworkNone
	}

	if strings.Contains(
		string(requirementsTxt), "django",
	) || utils.HasFile(absPath, "manage.py") {
		return PythonFrameworkDjango
	}

	if strings.Contains(string(requirementsTxt), "flask") {
		return PythonFrameworkFlask
	}

	return PythonFrameworkNone
}

func DetermineEntry(absPath string) string {
	if utils.HasFile(absPath, "main.py") {
		return "main.py"
	}

	if utils.HasFile(absPath, "app.py") {
		return "app.py"
	}

	if utils.HasFile(absPath, "manage.py") {
		return "manage.py"
	}

	return ""
}

func DetermineDependencyPolicy(absPath string) string {
	if utils.HasFile(absPath, "requirements.txt") {
		return "requirements.txt"
	}

	if utils.HasFile(absPath, "Pipfile") {
		return "Pipfile"
	}

	if utils.HasFile(absPath, "pyproject.toml") {
		return "pyproject.toml"
	}

	if utils.HasFile(absPath, "poetry.lock") {
		return "poetry.lock"
	}

	return ""
}
