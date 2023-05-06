package python

import (
	"regexp"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

type pythonPlanContext struct {
	Src            afero.Fs
	DependencyFile optional.Option[string]
	Framework      optional.Option[PythonFramework]
	Entry          optional.Option[string]
	Wsgi           optional.Option[string]
}

func DetermineFramework(ctx *pythonPlanContext) PythonFramework {
	src := ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	requirementsTxt, err := afero.ReadFile(src, "requirements.txt")
	if err != nil {
		*fw = optional.Some(PythonFrameworkNone)
		return fw.Unwrap()
	}

	req := string(requirementsTxt)
	if utils.WeakContains(req, "django") {
		*fw = optional.Some(PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.HasFile(src, "manage.py") {
		*fw = optional.Some(PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.WeakContains(req, "flask") {
		*fw = optional.Some(PythonFrameworkFlask)
		return fw.Unwrap()
	}

	*fw = optional.Some(PythonFrameworkNone)
	return fw.Unwrap()
}

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

func DetermineDependencyPolicy(ctx *pythonPlanContext) string {
	src := ctx.Src
	df := &ctx.DependencyFile

	if depFile, err := df.Take(); err == nil {
		return depFile
	}

	for _, file := range []string{"requirements.txt", "Pipfile", "pyproject.toml"} {
		if utils.HasFile(src, file) {
			*df = optional.Some(file)
			return df.Unwrap()
		}
	}

	*df = optional.Some("requirements.txt")
	return df.Unwrap()
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

func DetermineWsgi(ctx *pythonPlanContext) string {
	src := ctx.Src
	wa := &ctx.Wsgi

	framework := DetermineFramework(ctx)

	if framework == PythonFrameworkDjango {

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

	if framework == PythonFrameworkFlask {
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

	return ""
}

func determineInstallCmd(ctx *pythonPlanContext) string {
	depPolicy := DetermineDependencyPolicy(ctx)
	wsgi := DetermineWsgi(ctx)

	switch depPolicy {
	case "requirements.txt":
		if wsgi != "" {
			return "pip install -r requirements.txt && pip install gunicorn"
		} else {
			return "pip install -r requirements.txt"
		}
	case "Pipfile":
		if wsgi != "" {
			return "pipenv install && pipenv install gunicorn"
		} else {
			return "pipenv install"
		}
	case "pyproject.toml":
		if wsgi != "" {
			return "poetry install && poetry install gunicorn"
		} else {
			return "poetry install"
		}
	default:
		if wsgi != "" {
			return "pip install gunicorn"
		} else {
			return "echo \"skip install\""
		}
	}
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
	if wsgi != "" {
		return "gunicorn --bind :8080 " + wsgi
	}

	entry := DetermineEntry(ctx)
	return "python " + entry
}

type GetMetaOptions struct {
	Src afero.Fs
}

func GetMeta(opt GetMetaOptions) PlanMeta {
	ctx := &pythonPlanContext{Src: opt.Src}

	meta := PlanMeta{}

	framework := DetermineFramework(ctx)
	if framework != PythonFrameworkNone {
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
