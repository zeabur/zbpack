package python

import (
	"github.com/moznion/go-optional"
	"github.com/zeabur/zbpack/internal/source"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
	"regexp"
	"strings"
)

type pythonPlanContext struct {
	Src            *source.Source
	DependencyFile optional.Option[string]
	Framework      optional.Option[PythonFramework]
	Entry          optional.Option[string]
	Wsgi           optional.Option[string]
}

func DetermineFramework(ctx *pythonPlanContext) PythonFramework {
	src := *ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	requirementsTxt, err := src.ReadFile("requirements.txt")
	if err != nil {
		*fw = optional.Some(PythonFrameworkNone)
		return fw.Unwrap()
	}

	req := string(requirementsTxt)
	if utils.WeakContains(req, "django") {
		*fw = optional.Some(PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if src.HasFile("manage.py") {
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
	src := *ctx.Src
	et := &ctx.Entry

	if entry, err := et.Take(); err == nil {
		return entry
	}

	for _, file := range []string{"main.py", "app.py", "manage.py"} {
		if src.HasFile(file) {
			*et = optional.Some(file)
			return et.Unwrap()
		}
	}

	*et = optional.Some("main.py")
	return et.Unwrap()
}

func DetermineDependencyPolicy(ctx *pythonPlanContext) string {
	src := *ctx.Src
	df := &ctx.DependencyFile

	if depFile, err := df.Take(); err == nil {
		return depFile
	}

	for _, file := range []string{"requirements.txt", "Pipfile", "pyproject.toml", "poetry.lock"} {
		if src.HasFile(file) {
			*df = optional.Some(file)
			return df.Unwrap()
		}
	}

	*df = optional.Some("requirements.txt")
	return df.Unwrap()
}

func DetermineWsgi(ctx *pythonPlanContext) string {
	src := *ctx.Src
	wa := &ctx.Wsgi

	framework := DetermineFramework(ctx)

	if framework == PythonFrameworkDjango {

		dir, err := src.ReadDir("/")
		if err != nil {
			return ""
		}

		for _, d := range dir {
			if d.IsDir {
				if src.HasFile(d.Name + "/wsgi.py") {
					*wa = optional.Some(d.Name + ".wsgi")
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
		content, err := src.ReadFile(entryFile)
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
	case "poetry.lock":
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

func determineNeedMySQL(ctx *pythonPlanContext) bool {
	src := *ctx.Src

	p := DetermineDependencyPolicy(ctx)
	file, err := src.ReadFile(p)
	if err != nil {
		return false
	}

	if strings.Contains(string(file), "mysqlclient") {
		return true
	}

	// it probably doesn't have a dependency on `mysqlclient`
	return false
}

func determineNeedPostgreSQL(ctx *pythonPlanContext) bool {
	src := *ctx.Src

	p := DetermineDependencyPolicy(ctx)
	file, err := src.ReadFile(p)
	if err != nil {
		return false
	}

	if strings.Contains(string(file), "psycopg2") {
		return true
	}

	return false
}

func determineAptDependencies(ctx *pythonPlanContext) []string {
	needMySQL := determineNeedMySQL(ctx)
	needPostgreSQL := determineNeedPostgreSQL(ctx)

	if needMySQL {
		return []string{"libmariadb-dev", "build-essential"}
	}

	if needPostgreSQL {
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
	Src *source.Source
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
