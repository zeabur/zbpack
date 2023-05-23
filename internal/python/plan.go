// Package python is the build planner for Python projects.
package python

import (
	"regexp"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

type pythonPlanContext struct {
	Src            afero.Fs
	PackageManager optional.Option[types.PackageManager]
	Framework      optional.Option[types.PythonFramework]
	Entry          optional.Option[string]
	Wsgi           optional.Option[string]
}

// DetermineFramework determines the framework of the Python project.
func DetermineFramework(ctx *pythonPlanContext) types.PythonFramework {
	src := ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if HasDependency(src, "django") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.HasFile(src, "manage.py") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if HasDependency(src, "flask") {
		*fw = optional.Some(types.PythonFrameworkFlask)
		return fw.Unwrap()
	}

	if HasDependency(src, "fastapi") {
		*fw = optional.Some(types.PythonFrameworkFastapi)
		return fw.Unwrap()
	}

	*fw = optional.Some(types.PythonFrameworkNone)
	return fw.Unwrap()
}

// DetermineEntry determines the entry of the Python project.
func DetermineEntry(ctx *pythonPlanContext) string {
	src := ctx.Src
	et := &ctx.Entry

	if entry, err := et.Take(); err == nil {
		return entry
	}

	for _, file := range []string{"main.py", "app.py", "manage.py"} {
		if utils.HasFile(src, file) {
			*et = optional.Some(file)
			return et.Unwrap()
		}
	}

	*et = optional.Some("main.py")
	return et.Unwrap()
}

// DeterminePackageManager determines the package manager of this Python project.
func DeterminePackageManager(ctx *pythonPlanContext) types.PackageManager {
	src := ctx.Src
	cpm := &ctx.PackageManager

	// Pipfile > pyproject.toml > requirements.txt
	depFile := map[types.PackageManager]string{
		types.PythonPackageManagerPipenv: "Pipfile",
		types.PythonPackageManagerPoetry: "pyproject.toml",
		types.PythonPackageManagerPip:    "requirements.txt",
	}

	if packageManager, err := cpm.Take(); err == nil {
		return packageManager
	}

	for pm, file := range depFile {
		if utils.HasFile(src, file) {
			*cpm = optional.Some(pm)
			return cpm.Unwrap()
		}
	}

	*cpm = optional.Some(types.PythonPackageManagerUnknown)
	return cpm.Unwrap()
}

// HasDependency checks if a python project has the one of the dependencies.
func HasDependency(src afero.Fs, dependencies ...string) bool {
	for _, file := range []string{"requirements.txt", "Pipfile", "pyproject.toml", "Pipfile.lock", "poetry.lock"} {
		file, err := afero.ReadFile(src, file)
		if err != nil {
			continue
		}

		for _, dependency := range dependencies {
			if strings.Contains(string(file), dependency) {
				return true
			}
		}
	}

	return false
}

// DetermineWsgi determines the WSGI application filepath of a Python project.
func DetermineWsgi(ctx *pythonPlanContext) string {
	src := ctx.Src
	wa := &ctx.Wsgi

	if wsgi, err := wa.Take(); err == nil {
		return wsgi
	}

	framework := DetermineFramework(ctx)

	if framework == types.PythonFrameworkDjango {

		dir, err := afero.ReadDir(src, "/")
		if err != nil {
			return ""
		}

		for _, d := range dir {
			if d.IsDir() {
				if utils.HasFile(src, d.Name()+"/wsgi.py") {
					*wa = optional.Some(d.Name() + ".wsgi")
					return wa.Unwrap()
				}
			}
		}

		return ""
	}

	if framework == types.PythonFrameworkFlask {
		entryFile := DetermineEntry(ctx)
		// if there is something like `app = Flask(__name__)` in the entry file
		// we use this variable (app) as the wsgi application
		re := regexp.MustCompile(`(\w+)\s*=\s*Flask\([^)]*\)`)
		content, err := afero.ReadFile(src, entryFile)
		if err != nil {
			return ""
		}

		match := re.FindStringSubmatch(string(content))
		if len(match) > 1 {
			entryWithoutExt := strings.Replace(entryFile, ".py", "", 1)
			*wa = optional.Some(entryWithoutExt + ":" + match[1])
			return wa.Unwrap()
		}

		return ""
	}

	if framework == types.PythonFrameworkFastapi {
		entryFile := DetermineEntry(ctx)
		// if there is something like `app = FastAPI(__name__)` in the entry file
		// we use this variable (app) as the wsgi application
		re := regexp.MustCompile(`(\w+)\s*=\s*FastAPI\([^)]*\)`)
		content, err := afero.ReadFile(src, entryFile)
		if err != nil {
			return ""
		}

		match := re.FindStringSubmatch(string(content))
		if len(match) > 1 {
			entryWithoutExt := strings.Replace(entryFile, ".py", "", 1)
			*wa = optional.Some(entryWithoutExt + ":" + match[1])
			return wa.Unwrap()
		}

		return ""
	}

	return ""
}

func determineInstallCmd(ctx *pythonPlanContext) string {
	pm := DeterminePackageManager(ctx)
	wsgi := DetermineWsgi(ctx)
	framework := DetermineFramework(ctx)

	// Will be joined with `&&`
	andCommands := []string{}

	switch pm {
	case types.PythonPackageManagerPipenv:
		andCommands = append(andCommands, "pip install pipenv", "pipenv install")

		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				andCommands = append(andCommands, "pipenv install uvicorn")
			} else {
				andCommands = append(andCommands, "pipenv install gunicorn")
			}
		}
	case types.PythonPackageManagerPoetry:
		andCommands = append(andCommands, "pip install poetry", "poetry install")

		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				andCommands = append(andCommands, "poetry add uvicorn")
			} else {
				andCommands = append(andCommands, "poetry add gunicorn")
			}
		}
	case types.PythonPackageManagerPip:
		andCommands = append(andCommands, "pip install -r requirements.txt")
		fallthrough
	default:
		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				andCommands = append(andCommands, "pip install uvicorn")
			} else {
				andCommands = append(andCommands, "pip install gunicorn")
			}
		}
	}

	command := strings.Join(andCommands, " && ")
	if command != "" {
		return command
	}
	return "echo \"skip install\""
}

func determineAptDependencies(ctx *pythonPlanContext) []string {
	if HasDependency(ctx.Src, "mysqlclient") {
		return []string{"libmariadb-dev", "build-essential"}
	}

	if HasDependency(ctx.Src, "psycopg2") {
		return []string{"libpq-dev"}
	}

	return []string{}
}

func determineStartCmd(ctx *pythonPlanContext) string {
	wsgi := DetermineWsgi(ctx)
	framework := DetermineFramework(ctx)

	var commandSegment []string

	switch DeterminePackageManager(ctx) {
	case types.PythonPackageManagerPipenv:
		commandSegment = append(commandSegment, "pipenv run")
	case types.PythonPackageManagerPoetry:
		commandSegment = append(commandSegment, "poetry run")
	}

	if wsgi != "" {
		if framework == types.PythonFrameworkFastapi {
			commandSegment = append(commandSegment, "uvicorn", wsgi, "--host 0.0.0.0", "--port 8080")
		} else {
			commandSegment = append(commandSegment, "gunicorn", "--bind :8080", wsgi)
		}
	} else {
		entry := DetermineEntry(ctx)
		commandSegment = append(commandSegment, "python", entry)
	}

	command := strings.Join(commandSegment, " ")
	return command
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src afero.Fs
}

// GetMeta returns the metadata of a Python project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := &pythonPlanContext{Src: opt.Src}

	meta := types.PlanMeta{}

	framework := DetermineFramework(ctx)
	if framework != types.PythonFrameworkNone {
		meta["framework"] = string(framework)
	}

	installCmd := determineInstallCmd(ctx)
	meta["install"] = installCmd

	startCmd := determineStartCmd(ctx)
	meta["start"] = startCmd

	aptDeps := determineAptDependencies(ctx)
	if len(aptDeps) > 0 {
		meta["apt-deps"] = strings.Join(aptDeps, " ")
	}

	return meta
}
