// Package python is the build planner for Python projects.
package python

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
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
	Static         optional.Option[StaticInfo]
}

// DetermineFramework determines the framework of the Python project.
func DetermineFramework(ctx *pythonPlanContext) types.PythonFramework {
	src := ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if HasDependencyWithFile(ctx, "django") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.HasFile(src, "manage.py") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if HasDependencyWithFile(ctx, "flask") {
		*fw = optional.Some(types.PythonFrameworkFlask)
		return fw.Unwrap()
	}

	if HasDependencyWithFile(ctx, "fastapi") {
		*fw = optional.Some(types.PythonFrameworkFastapi)
		return fw.Unwrap()
	}

	if HasDependencyWithFile(ctx, "sanic") {
		*fw = optional.Some(types.PythonFrameworkSanic)
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

	for _, file := range []string{"main.py", "app.py", "manage.py", "server.py"} {
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
	depFiles := []struct {
		packageManagerID types.PackageManager
		filename         string
		content          string
		lockFile         string
	}{
		{types.PythonPackageManagerPipenv, "Pipfile", "", ""},
		{types.PythonPackageManagerPoetry, "pyproject.toml", "[tool.poetry]", "poetry.lock"},
		{types.PythonPackageManagerPdm, "pyproject.toml", "[tool.pdm]", "pdm.lock"},
		{types.PythonPackageManagerPip, "requirements.txt", "", ""},
	}

	if packageManager, err := cpm.Take(); err == nil {
		return packageManager
	}

	for _, depFile := range depFiles {
		if depFile.content == "" && depFile.lockFile == "" {
			if utils.HasFile(src, depFile.filename) {
				*cpm = optional.Some(depFile.packageManagerID)
				return cpm.Unwrap()
			}
		} else if depFile.content != "" && depFile.lockFile == "" {
			if utils.HasFile(src, depFile.filename) && weakHasStringsInFiles(src, []string{depFile.filename}, depFile.content) {
				*cpm = optional.Some(depFile.packageManagerID)
				return cpm.Unwrap()
			}
		} else if depFile.content != "" && depFile.lockFile != "" {
			if utils.HasFile(src, depFile.filename) {
				if weakHasStringsInFiles(src, []string{depFile.filename}, depFile.content) || utils.HasFile(src, depFile.lockFile) {
					*cpm = optional.Some(depFile.packageManagerID)
					return cpm.Unwrap()
				}
			}
		}
	}

	*cpm = optional.Some(types.PythonPackageManagerUnknown)
	return cpm.Unwrap()
}

// HasDependency checks if the specified dependency is in the project.
func HasDependency(ctx *pythonPlanContext, dependency string) bool {
	src := ctx.Src
	pm := DeterminePackageManager(ctx)

	switch pm {
	case types.PythonPackageManagerPip:
		return weakHasStringsInFiles(src, []string{"requirements.txt"}, dependency)
	case types.PythonPackageManagerPoetry:
		return weakHasStringsInFiles(src, []string{"pyproject.toml", "poetry.lock"}, dependency)
	case types.PythonPackageManagerPipenv:
		return weakHasStringsInFiles(src, []string{"Pipfile", "Pipfile.lock"}, dependency)
	case types.PythonPackageManagerPdm:
		return weakHasStringsInFiles(src, []string{"pyproject.toml", "pdm.lock"}, dependency)
	}

	return false
}

// weakHasStringsInFiles checks if the specified text are in the listed files.
func weakHasStringsInFiles(src afero.Fs, filelist []string, text string) bool {
	for _, file := range filelist {
		file, err := afero.ReadFile(src, file)
		if err != nil {
			continue
		}

		if utils.WeakContains(string(file), text) {
			return true
		}
	}

	return false
}

// HasDependencyWithFile checks if the specified dependency is in the file.
func HasDependencyWithFile(ctx *pythonPlanContext, dependency string) bool {
	src := ctx.Src
	pm := DeterminePackageManager(ctx)

	switch pm {
	case types.PythonPackageManagerPip:
		return weakHasStringsInFile(src, "requirements.txt", dependency)
	case types.PythonPackageManagerPipenv:
		return weakHasStringsInFile(src, "Pipfile", dependency)
	case types.PythonPackageManagerPoetry:
		return weakHasStringsInFile(src, "pyproject.toml", dependency)
	case types.PythonPackageManagerPdm:
		return weakHasStringsInFile(src, "pyproject.toml", dependency)
	}

	return false
}

// weakHasStringsInFile checks if the specified text are in the file.
func weakHasStringsInFile(src afero.Fs, file string, text string) bool {
	content, err := afero.ReadFile(src, file)
	if err != nil {
		return false
	}

	if utils.WeakContains(string(content), text) {
		return true
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

	if framework == types.PythonFrameworkSanic {
		entryFile := DetermineEntry(ctx)

		re := regexp.MustCompile(`(\w+)\s*=\s*Sanic\([^)]*\)`)
		content, err := afero.ReadFile(src, entryFile)
		if err != nil {
			return ""
		}

		match := re.FindStringSubmatch(string(content))
		if len(match) > 1 {
			entryWithoutExt := strings.TrimSuffix(entryFile, ".py")
			*wa = optional.Some(entryWithoutExt + ":" + match[1])
			return wa.Unwrap()
		}
		return ""
	}

	return ""
}

// getDjangoSettings finds and reads the Django settings module
// of a Python project.
func getDjangoSettings(fs afero.Fs) ([]byte, error) {
	djangoSettingModule := regexp.MustCompile(`['"]DJANGO_SETTINGS_MODULE['"],\s*['"](.+)\.settings['"]`)

	// According to https://github.com/django/django/blob/bcd80de8b5264d8c
	// 853bbd38bfeb02279a9b3799/django/conf/__init__.py#L61, it reads
	// the "DJANGO_SETTINGS_MODULE" environment variable to determine
	// where the settings module is.

	// Generally, the "DJANGO_SETTINGS_MODULE" environment variable
	// is defined in the "manage.py" file. So we read the manage.py first.
	managePy, err := afero.ReadFile(fs, "manage.py")
	if err != nil {
		return nil, fmt.Errorf("read manage.py: %w", err)
	}

	// We try to find the line defining the "DJANGO_SETTINGS_MODULE"
	// environment variable. The line is usually like:
	//
	//		os.environ.setdefault("DJANGO_SETTINGS_MODULE", "myproject.settings")
	//
	// We try to find the "myproject" substring here.
	match := djangoSettingModule.FindSubmatch(managePy)
	if len(match) != 2 {
		// We should only have one match.
		return nil, errors.New("no DJANGO_SETTINGS_MODULE defined")
	}

	// We try to read the settings.py file declaring in the
	// "DJANGO_SETTINGS_MODULE" environment variable.
	settingsFile, err := afero.ReadFile(fs, filepath.Join(string(match[1]), "settings.py"))
	if err != nil {
		return nil, fmt.Errorf("read settings.py: %w", err)
	}

	// Found!
	return settingsFile, nil
}

// DetermineStaticInfo determines the static path for Nginx to host.
// If this returns "", it means that we don't need to host static files
// with Nginx; otherwise, it returns the path to the static files.
func DetermineStaticInfo(ctx *pythonPlanContext) StaticInfo {
	var (
		// staticURLRegex matches the following:
		//
		//    STATIC_URL = '</static/>' ($2)
		//    STATIC_URL="</staticexample>" ($2)
		staticURLRegex = regexp.MustCompile(`STATIC_URL\s*=\s*['"]([^'"]*)['"]`)
		// staticRootRegex matches the following:
		//
		//   STATIC_ROOT = os.path.join(BASE_DIR, "<staticfiles>") ($2)
		//   STATIC_ROOT = BASE_DIR / "<staticfiles>" ($3)
		staticRootRegex      = regexp.MustCompile(`STATIC_ROOT\s*=\s*(?:os.path.join\(BASE_DIR,\s*["'](.+)["']\)|BASE_DIR\s*/\s*["'](.+)["'])`)
		staticURLCheckRegex  = regexp.MustCompile(`STATIC_URL\s*=`)
		staticRootCheckRegex = regexp.MustCompile(`STATIC_ROOT\s*=`)
	)

	const defaultStaticURL = "/static/"
	const defaultDjangoBaseDir = "/app/"
	const defaultDjangoStaticHostDir = defaultDjangoBaseDir + "staticfiles/"

	src := ctx.Src
	sp := &ctx.Static

	if staticInfo, err := sp.Take(); err == nil {
		return staticInfo
	}

	framework := DetermineFramework(ctx)

	if framework == types.PythonFrameworkDjango {
		settings, err := getDjangoSettings(src)
		if err != nil {
			// Assuming this project does not enable static file.
			log.Println("getDjangoSettings:", err)

			*sp = optional.Some(StaticInfo{})
			return sp.Unwrap()
		}

		if staticRootCheckRegex.Match(settings) && staticURLCheckRegex.Match(settings) {
			// We don't need to start an additional nginx server if user
			// has specified "whitenoise.middleware.WhiteNoiseMiddleware"
			// middleware. FIXME: we don't check if the middleware is
			// actually enabled (for example, commented out.)
			if strings.Contains(string(settings), "whitenoise.middleware.WhiteNoiseMiddleware") {
				*sp = optional.Some(StaticInfo{
					Flag: StaticModeDjango,
				})
				return sp.Unwrap()
			}

			// Add "/" prefix to the static url path if it doesn't have one.
			staticURLPath := defaultStaticURL
			if match := staticURLRegex.FindSubmatch(settings); len(match) > 1 {
				staticURLPath = string(match[1])

				if !strings.HasPrefix(staticURLPath, "/") {
					staticURLPath = "/" + staticURLPath
				}
			}

			// Find the static root
			staticRootPath := defaultDjangoStaticHostDir
			if match := staticRootRegex.FindSubmatch(settings); len(match) > 1 {
				// find the first non-empty match
				for _, m := range match[1:] {
					if len(m) > 0 {
						staticRootPath = defaultDjangoBaseDir + string(m)
					}
				}

				// add "/" suffix to the static root if it doesn't have one
				if !strings.HasSuffix(staticRootPath, "/") {
					staticRootPath = staticRootPath + "/"
				}
			}

			// Otherwise, we need to host static files with Nginx.
			*sp = optional.Some(StaticInfo{
				Flag:          StaticModeDjango | StaticModeNginx,
				StaticURLPath: staticURLPath,
				StaticHostDir: staticRootPath,
			})
			return sp.Unwrap()
		}

		// For any other configuration of Django (including none),
		// we assume that we don't need to host static files.
		*sp = optional.Some(StaticInfo{})
		return sp.Unwrap()
	}

	// For any other framework (including none), we assume that we don't
	// need to host static files.
	*sp = optional.Some(StaticInfo{})
	return sp.Unwrap()
}

func determineInstallCmd(ctx *pythonPlanContext) string {
	pm := DeterminePackageManager(ctx)
	wsgi := DetermineWsgi(ctx)
	framework := DetermineFramework(ctx)

	// will be joined by newline
	var commands []string

	switch pm {
	case types.PythonPackageManagerPipenv:
		commands = append(commands, "RUN pip install pipenv")

		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				commands = append(commands, "RUN pipenv install uvicorn")
			} else {
				commands = append(commands, "RUN pipenv install gunicorn")
			}
		}
		commands = append(commands, "COPY Pipfile* .", "RUN pipenv install")
	case types.PythonPackageManagerPoetry:
		commands = append(commands, "RUN pip install poetry")

		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				commands = append(commands, "RUN poetry add uvicorn")
			} else {
				commands = append(commands, "RUN poetry add gunicorn")
			}
		}
		commands = append(commands, "COPY poetry.lock* pyproject.toml* .", "RUN poetry install")
	case types.PythonPackageManagerPdm:
		commands = append(commands, "COPY pdm.lock* pyproject.toml* .", "RUN pip install pdm")
		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				commands = append(commands, "RUN pdm add uvicorn")
			} else {
				commands = append(commands, "RUN pdm add gunicorn")
			}
		}
		commands = append(commands, "RUN pdm install")
	case types.PythonPackageManagerPip:
		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				commands = append(commands, "RUN pip install uvicorn")
			} else {
				commands = append(commands, "RUN pip install gunicorn")
			}
		}
		commands = append(commands, "COPY requirements.txt* .", "RUN pip install -r requirements.txt")
	default:
		if wsgi != "" {
			if framework == types.PythonFrameworkFastapi {
				commands = append(commands, "RUN pip install uvicorn")
			} else {
				commands = append(commands, "RUN pip install gunicorn")
			}
		}
	}

	command := strings.Join(commands, "\n")
	if command != "" {
		return command
	}
	return "RUN echo \"skip install\""
}

func determineAptDependencies(ctx *pythonPlanContext) []string {
	deps := []string{"build-essential", "pkg-config"}

	// If we need to host static files, we need nginx.
	staticPath := DetermineStaticInfo(ctx)
	if staticPath.NginxEnabled() {
		deps = append(deps, "nginx")
	}

	if HasDependency(ctx, "mysqlclient") {
		deps = append(deps, "libmariadb-dev")
	}

	if HasDependency(ctx, "psycopg2") {
		deps = append(deps, "libpq-dev")
	}

	if HasDependency(ctx, "pyzbar") {
		deps = append(deps, "libzbar0")
	}

	if HasDependency(ctx, "chromadb") {
		deps = append(deps, "g++-7")
	}

	return deps
}

func determineStartCmd(ctx *pythonPlanContext) string {
	wsgi := DetermineWsgi(ctx)
	framework := DetermineFramework(ctx)
	pm := DeterminePackageManager(ctx)
	staticPath := DetermineStaticInfo(ctx)

	var commandSegment []string

	// We need Nginx server if we need to host static files.
	if staticPath.NginxEnabled() {
		commandSegment = append(commandSegment, "/usr/sbin/nginx &&")
	}

	switch pm {
	case types.PythonPackageManagerPipenv:
		commandSegment = append(commandSegment, "pipenv run")
	case types.PythonPackageManagerPoetry:
		commandSegment = append(commandSegment, "poetry run")
	case types.PythonPackageManagerPdm:
		commandSegment = append(commandSegment, "pdm run")
	}

	if wsgi != "" {
		wsgilistenedPort := "8080"

		// The WSGI application should listen at 8000
		// for reverse proxying by Nginx if we need to
		// host static files with Nginx. The "8000" is
		// configured by our nginx.conf in `python.go`.
		if staticPath.NginxEnabled() {
			wsgilistenedPort = "8000"
		}

		if framework == types.PythonFrameworkFastapi {
			commandSegment = append(commandSegment, "uvicorn", wsgi, "--host 0.0.0.0", "--port "+wsgilistenedPort)
		} else if framework == types.PythonFrameworkSanic {
			commandSegment = append(commandSegment, "sanic", wsgi, "--host 0.0.0.0", "--port "+wsgilistenedPort)
		} else {
			commandSegment = append(commandSegment, "gunicorn", "--bind :"+wsgilistenedPort, wsgi)
		}
	} else {
		entry := DetermineEntry(ctx)
		commandSegment = append(commandSegment, "python", entry)
	}

	command := strings.Join(commandSegment, " ")
	return command
}

// determinePythonVersion Determine Python Version
func determinePythonVersion(ctx *pythonPlanContext) string {
	pm := DeterminePackageManager(ctx)

	switch pm {
	case types.PythonPackageManagerPoetry:
		return determinePythonVersionWithPoetry(ctx)
	case types.PythonPackageManagerPdm:
		return determinePythonVersionWithPdm(ctx)
	default:
		return defaultPython3Version
	}
}

func determinePythonVersionWithPdm(ctx *pythonPlanContext) string {
	src := ctx.Src

	content, err := afero.ReadFile(src, "pyproject.toml")
	if err != nil {
		return defaultPython3Version
	}

	compile := regexp.MustCompile(`requires-python = "(.*?)"`)
	submatchs := compile.FindStringSubmatch(string(content))
	if len(submatchs) > 1 {
		version := submatchs[1]
		return getPython3Version(version)
	}

	return defaultPython3Version
}

func determinePythonVersionWithPoetry(ctx *pythonPlanContext) string {
	src := ctx.Src

	content, err := afero.ReadFile(src, "pyproject.toml")
	if err != nil {
		return defaultPython3Version
	}

	compile := regexp.MustCompile(`python = "(.*?)"`)
	submatchs := compile.FindStringSubmatch(string(content))
	if len(submatchs) > 1 {
		version := submatchs[1]
		return getPython3Version(version)
	}

	return defaultPython3Version
}

func determineBuildCmd(ctx *pythonPlanContext) string {
	staticInfo := DetermineStaticInfo(ctx)

	if staticInfo.DjangoEnabled() {
		// We need to collect static files if we are using Django.
		return "RUN python manage.py collectstatic --noinput"
	}

	return ""
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src afero.Fs
}

// GetMeta returns the metadata of a Python project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	meta := types.PlanMeta{}

	ctx := &pythonPlanContext{
		Src: opt.Src,
	}

	pm := DeterminePackageManager(ctx)
	meta["packageManager"] = string(pm)

	version := determinePythonVersion(ctx)
	meta["pythonVersion"] = version

	staticPath := DetermineStaticInfo(ctx)
	for k, v := range staticPath.Meta() {
		meta[k] = v
	}

	framework := DetermineFramework(ctx)
	if framework != types.PythonFrameworkNone {
		meta["framework"] = string(framework)
	}

	installCmd := determineInstallCmd(ctx)
	meta["install"] = installCmd

	buildCmd := determineBuildCmd(ctx)
	if buildCmd != "" {
		meta["build"] = buildCmd
	}

	startCmd := determineStartCmd(ctx)
	meta["start"] = startCmd

	aptDeps := determineAptDependencies(ctx)
	if len(aptDeps) > 0 {
		meta["apt-deps"] = strings.Join(aptDeps, " ")
	}

	return meta
}
