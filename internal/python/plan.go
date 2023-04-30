package python

import (
	"bufio"
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
	"regexp"
	"strings"
)

type pythonPlanContext struct {
	SrcFs          afero.Fs
	DependencyFile optional.Option[string]
	Framework      optional.Option[PythonFramework]
	Entry          optional.Option[string]
	Wsgi           optional.Option[string]
}

func DetermineFramework(ctx *pythonPlanContext) PythonFramework {
	fs := ctx.SrcFs
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	requirementsTxt, err := afero.ReadFile(fs, "requirements.txt")
	if err != nil {
		*fw = optional.Some(PythonFrameworkNone)
		return fw.Unwrap()
	}

	req := string(requirementsTxt)
	if utils.Contains(req, "django") {
		*fw = optional.Some(PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if _, err := fs.Stat("manage.py"); err == nil {
		*fw = optional.Some(PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.Contains(req, "flask") {
		*fw = optional.Some(PythonFrameworkFlask)
		return fw.Unwrap()
	}

	*fw = optional.Some(PythonFrameworkNone)
	return fw.Unwrap()
}

func DetermineEntry(ctx *pythonPlanContext) string {
	fs := ctx.SrcFs
	et := &ctx.Entry

	if entry, err := et.Take(); err == nil {
		return entry
	}

	for _, file := range []string{"main.py", "app.py", "manage.py"} {
		if _, err := fs.Stat(file); err == nil {
			*et = optional.Some(file)
			return et.Unwrap()
		}
	}

	*et = optional.Some("main.py")
	return et.Unwrap()
}

func DetermineDependencyPolicy(ctx *pythonPlanContext) string {
	fs := ctx.SrcFs
	df := &ctx.DependencyFile

	if depFile, err := df.Take(); err == nil {
		return depFile
	}

	for _, file := range []string{"requirements.txt", "Pipfile", "pyproject.toml", "poetry.lock"} {
		if _, err := fs.Stat(file); err == nil {
			*df = optional.Some(file)
			return df.Unwrap()
		}
	}

	*df = optional.Some("requirements.txt")
	return df.Unwrap()
}

func DetermineWsgi(ctx *pythonPlanContext) string {
	fs := ctx.SrcFs
	wa := &ctx.Wsgi

	framework := DetermineFramework(ctx)

	if framework == PythonFrameworkDjango {

		dir, err := afero.ReadDir(fs, "/")
		if err != nil {
			return ""
		}

		for _, d := range dir {
			if d.IsDir() {
				if _, err := fs.Stat("/" + d.Name() + "/wsgi.py"); err == nil {
					*wa = optional.Some(d.Name() + ".wsgi")
					return wa.Unwrap()
				}
			}
		}

		return ""
	}

	if framework == PythonFrameworkFlask {
		entryFile := DetermineEntry(ctx)
		re := regexp.MustCompile(`(\w+)\s*=\s*Flask\([^)]*\)`)
		content, err := afero.ReadFile(fs, entryFile)
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
	fs := ctx.SrcFs

	possiblePath := []string{
		"requirements.txt",
		"pyproject.toml",
		"poetry.lock",
		"Pipfile",
		"Pipfile.lock",
	}

	for _, p := range possiblePath {
		file, err := fs.Open(p)
		if err != nil {
			// the file may not exist, and we can safely ignore this error
			continue
		}
		defer file.Close()

		// read file line by line – usually the string `mysqlclient`
		// will be present completely on one line.
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "mysqlclient") {
				return true
			}
		}
	}

	// it probably doesn't have a dependency on `mysqlclient`
	return false
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
	AbsPath string
}

func GetMeta(opt GetMetaOptions) PlanMeta {
	fs := afero.NewBasePathFs(afero.NewOsFs(), opt.AbsPath)

	ctx := &pythonPlanContext{SrcFs: fs}

	meta := PlanMeta{}

	framework := DetermineFramework(ctx)
	meta["framework"] = string(framework)

	installCmd := determineInstallCmd(ctx)
	meta["install"] = installCmd

	startCmd := determineStartCmd(ctx)
	meta["start"] = startCmd

	if determineNeedMySQL(ctx) {
		meta["needMySQL"] = "true"
	}

	return meta
}
