// Package python is the build planner for Python projects.
package python

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type pythonPlanContext struct {
	Src            afero.Fs
	Config         plan.ImmutableProjectConfiguration
	PackageManager optional.Option[types.PythonPackageManager]
	Framework      optional.Option[types.PythonFramework]
	Entry          optional.Option[string]
	Wsgi           optional.Option[string]
	Static         optional.Option[StaticInfo]
	StreamlitEntry optional.Option[string]
	Serverless     optional.Option[bool]
}

const (
	// ConfigStreamlitEntry is the key for specifying the streamlit entry explicitly
	// in the project configuration.
	ConfigStreamlitEntry = "streamlit.entry"

	// ConfigPythonEntry is the key for specifying the Python entry explicitly
	// in the project configuration. If there is `__init__.py`, you should also
	// write down the `__init__.py` in the entry. For example: `app/__init__.py`.
	ConfigPythonEntry = "python.entry"

	// ConfigPythonVersion is the key for specifying the Python version explicitly
	// in the project configuration.
	ConfigPythonVersion = "python.version"

	// ConfigPythonPackageManager is the key for specifying the Python package manager
	// explicitly in the project configuration.
	// Note that it should be one of the values in `types.PythonPackageManager`.
	ConfigPythonPackageManager = "python.package_manager"
)

// DetermineFramework determines the framework of the Python project.
func DetermineFramework(ctx *pythonPlanContext) types.PythonFramework {
	src := ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if HasExplicitDependency(ctx, "reflex") {
		*fw = optional.Some(types.PythonFrameworkReflex)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "django") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if utils.HasFile(src, "manage.py") {
		*fw = optional.Some(types.PythonFrameworkDjango)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "flask") {
		*fw = optional.Some(types.PythonFrameworkFlask)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "fastapi") {
		*fw = optional.Some(types.PythonFrameworkFastapi)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "tornado") {
		*fw = optional.Some(types.PythonFrameworkTornado)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "sanic") {
		*fw = optional.Some(types.PythonFrameworkSanic)
		return fw.Unwrap()
	}

	if HasExplicitDependency(ctx, "streamlit") {
		*fw = optional.Some(types.PythonFrameworkStreamlit)
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

	if entry, err := plan.Cast(ctx.Config.Get(ConfigPythonEntry), cast.ToStringE).Take(); err == nil {
		*et = optional.Some(entry)
		return et.Unwrap()
	}

	for _, file := range []string{"main.py", "app.py", "manage.py", "server.py", "app/__init__.py"} {
		if utils.HasFile(src, file) {
			*et = optional.Some(file)
			return et.Unwrap()
		}
	}

	*et = optional.Some("main.py")
	return et.Unwrap()
}

// DeterminePackageManager determines the package manager of this Python project.
func DeterminePackageManager(ctx *pythonPlanContext) types.PythonPackageManager {
	src := ctx.Src
	cpm := &ctx.PackageManager

	if packageManager, err := cpm.Take(); err == nil {
		return packageManager
	}

	/* User can specify the package manager explicitly */
	if packageManager, err := plan.Cast(ctx.Config.Get(ConfigPythonPackageManager), cast.ToStringE).Take(); err == nil {
		switch packageManager {
		case string(types.PythonPackageManagerPip),
			string(types.PythonPackageManagerPoetry),
			string(types.PythonPackageManagerPipenv),
			string(types.PythonPackageManagerPdm),
			string(types.PythonPackageManagerRye),
			string(types.PythonPackageManagerUv):
			*cpm = optional.Some(types.PythonPackageManager(packageManager))
			return cpm.Unwrap()
		default:
			*cpm = optional.Some(types.PythonPackageManagerUnknown)
			return cpm.Unwrap()
		}
	}

	/* Pipenv */
	// If there is a Pipfile, we use Pipenv.
	if utils.HasFile(src, "Pipfile") {
		*cpm = optional.Some(types.PythonPackageManagerPipenv)
		return cpm.Unwrap()
	}

	/* Poetry */
	// If there is poetry.lock, we use Poetry.
	if utils.HasFile(src, "poetry.lock") {
		*cpm = optional.Some(types.PythonPackageManagerPoetry)
		return cpm.Unwrap()
	}
	// If there is a pyproject.toml with [tool.poetry], we use Poetry.
	if content, err := utils.ReadFileToUTF8(src, "pyproject.toml"); err == nil && strings.Contains(string(content), "[tool.poetry]") {
		*cpm = optional.Some(types.PythonPackageManagerPoetry)
		return cpm.Unwrap()
	}

	/* Pdm */
	// If there is pdm.lock, we use Pdm.
	if utils.HasFile(src, "pdm.lock") {
		*cpm = optional.Some(types.PythonPackageManagerPdm)
		return cpm.Unwrap()
	}
	// If there is a pyproject.toml with [tool.pdm], we use Pdm.
	if content, err := utils.ReadFileToUTF8(src, "pyproject.toml"); err == nil && strings.Contains(string(content), "[tool.pdm]") {
		*cpm = optional.Some(types.PythonPackageManagerPdm)
		return cpm.Unwrap()
	}

	/* Pip */
	// If there is a requirements.txt, we use Pip.
	if utils.HasFile(src, "requirements.txt") {
		*cpm = optional.Some(types.PythonPackageManagerPip)
		return cpm.Unwrap()
	}

	/* Rye */
	// If there is a requirements.lock, we use Rye.
	if utils.HasFile(src, "requirements.lock") {
		*cpm = optional.Some(types.PythonPackageManagerRye)
		return cpm.Unwrap()
	}
	// If there is a pyproject.toml with [tool.rye], we use Rye.
	if content, err := utils.ReadFileToUTF8(src, "pyproject.toml"); err == nil && strings.Contains(string(content), "[tool.rye]") {
		*cpm = optional.Some(types.PythonPackageManagerRye)
		return cpm.Unwrap()
	}

	/* uv */
	// If there is a uv.lock, we use uv.
	if utils.HasFile(src, "uv.lock") {
		*cpm = optional.Some(types.PythonPackageManagerUv)
		return cpm.Unwrap()
	}

	*cpm = optional.Some(types.PythonPackageManagerUnknown)
	return cpm.Unwrap()
}

// HasDependency checks if the specified dependency is in the project.
func HasDependency(ctx *pythonPlanContext, dependency string) bool {
	src := ctx.Src
	pm := DeterminePackageManager(ctx)

	filesToFind := lo.Filter(
		append([]string{getPmDeclarationFile(pm)}, getPmLockFile(pm)...),
		func(s string, _ int) bool {
			return s != ""
		},
	)

	return weakHasStringsInFiles(src, filesToFind, dependency)
}

// weakHasStringsInFiles checks if the specified text are in the listed files.
func weakHasStringsInFiles(src afero.Fs, filelist []string, text string) bool {
	for _, file := range filelist {
		file, err := utils.ReadFileToUTF8(src, file)
		if err != nil {
			continue
		}

		if utils.WeakContains(string(file), text) {
			return true
		}
	}

	return false
}

// HasExplicitDependency checks if the specified dependency is specified explicitly in the project.
func HasExplicitDependency(ctx *pythonPlanContext, dependency string) bool {
	src := ctx.Src
	pm := DeterminePackageManager(ctx)

	if f := getPmDeclarationFile(pm); f != "" {
		if weakHasStringsInFile(src, f, dependency) {
			return true
		}
	}

	return false
}

// weakHasStringsInFile checks if the specified text are in the file.
func weakHasStringsInFile(src afero.Fs, file string, text string) bool {
	content, err := utils.ReadFileToUTF8(src, file)
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

	{
		// if there is something like `app = <Constructor>(__name__)` in the entry file
		// we use this variable (app) as the wsgi application
		constructor := ""
		switch framework {
		case types.PythonFrameworkFlask:
			constructor = "Flask"
		case types.PythonFrameworkFastapi:
			constructor = "FastAPI"
		case types.PythonFrameworkTornado:
			constructor = "Tornado"
		case types.PythonFrameworkSanic:
			constructor = "Sanic"
		}

		if constructor != "" {
			entryFile := DetermineEntry(ctx)

			re := regexp.MustCompile(`(\w+)\s*=\s*` + constructor + `\(`)
			content, err := utils.ReadFileToUTF8(src, entryFile)
			if err != nil {
				return ""
			}

			match := re.FindStringSubmatch(string(content))
			if len(match) > 1 {
				var finalEntry string

				// example/app/__init__.py -> example/app/__init__
				finalEntry = strings.TrimSuffix(entryFile, ".py")
				// example/app/__init__ -> example/app/
				finalEntry = strings.TrimSuffix(finalEntry, "__init__")
				// example/app/ -> example.app.
				finalEntry = strings.ReplaceAll(finalEntry, "/", ".")
				// example.app. -> example.app
				finalEntry = strings.TrimSuffix(finalEntry, ".")

				*wa = optional.Some(finalEntry + ":" + match[1])
				return wa.Unwrap()
			}
		}
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
	managePy, err := utils.ReadFileToUTF8(fs, "manage.py")
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
	settingsFile, err := utils.ReadFileToUTF8(fs, filepath.Join(string(match[1]), "settings.py"))
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
	serverless := getServerless(ctx)

	// will be joined by newline
	var commands []string

	var depToInstall []string
	if wsgi != "" && !serverless {
		if framework == types.PythonFrameworkFastapi {
			depToInstall = append(depToInstall, "uvicorn")
		} else {
			depToInstall = append(depToInstall, "gunicorn")
		}
	}

	if cmd := getPmInitCmd(pm); cmd != "" {
		commands = append(commands, "RUN "+cmd)
	}
	if cmd := getPmAddCmd(pm, depToInstall...); cmd != "" {
		commands = append(commands, "RUN "+cmd)
	}
	if cmd := getPmInstallCmd(pm); cmd != "" {
		commands = append(commands, "RUN "+cmd)
	}

	command := strings.Join(commands, "\n")
	if command != "" {
		return command
	}
	return "RUN echo \"skip install\""
}

func determineAptDependencies(ctx *pythonPlanContext) []string {
	serverless := getServerless(ctx)
	if serverless {
		return []string{}
	}

	framework := DetermineFramework(ctx)
	if framework == types.PythonFrameworkReflex {
		return []string{"caddy"}
	}

	deps := []string{"build-essential", "pkg-config", "clang"}

	// If we need to host static files, we need nginx.
	staticPath := DetermineStaticInfo(ctx)
	if staticPath.NginxEnabled() {
		deps = append(deps, "nginx")
	}

	if HasDependency(ctx, "mysqlclient") {
		deps = append(deps, "libmariadb-dev")
	}

	// psycopg2 included
	if HasDependency(ctx, "psycopg") {
		deps = append(deps, "libpq-dev")
	}

	if HasDependency(ctx, "pyzbar") {
		deps = append(deps, "libzbar0")
	}

	if HasDependency(ctx, "chromadb") {
		deps = append(deps, "g++-7")
	}

	if HasDependency(ctx, "pyaudio") {
		deps = append(deps, "portaudio19-dev")
	}

	if HasDependency(ctx, "azure-cognitiveservices-speech") {
		deps = append(deps, "libssl-dev", "libasound2")
	}

	if HasDependency(ctx, "moviepy") {
		deps = append(deps, "ffmpeg", "libsm6", "libxext6", "imagemagick", "cmake")
	}

	if determinePlaywright(ctx) {
		deps = append(
			deps, "libnss3", "libatk1.0-0", "libatk-bridge2.0-0",
			"libcups2", "libdbus-1-3", "libdrm2", "libxkbcommon-x11-0",
			"libxcomposite-dev", "libxdamage1", "libxfixes-dev", "libxrandr2",
			"libgbm-dev", "libasound2", "libpango-1.0-0", "libcairo-5c0",
		)
	}

	if ok, _ := afero.Exists(ctx.Src, "package.json"); ok {
		deps = append(deps, "nodejs", "npm")
	}

	return deps
}

func determineDefaultStartupFunction(ctx *pythonPlanContext) string {
	wsgi := DetermineWsgi(ctx)
	framework := DetermineFramework(ctx)
	pm := DeterminePackageManager(ctx)
	staticPath := DetermineStaticInfo(ctx)

	if framework == types.PythonFrameworkReflex {
		switch pm {
		case types.PythonPackageManagerPoetry:
			return "[ -d alembic ] && poetry run reflex db migrate; caddy start && poetry run reflex run --env prod --backend-only --loglevel debug"
		default:
			return "[ -d alembic ] && reflex db migrate; caddy start && reflex run --env prod --backend-only --loglevel debug"
		}
	}

	var commandSegment []string

	// We need Nginx server if we need to host static files.
	if staticPath.NginxEnabled() {
		commandSegment = append(commandSegment, "/usr/sbin/nginx &&")
	}

	if prefix := getPmStartCmdPrefix(pm); prefix != "" {
		commandSegment = append(commandSegment, prefix)
	}

	if streamlitEntry := determineStreamlitEntry(ctx); streamlitEntry != "" {
		commandSegment = append(commandSegment, "streamlit run", streamlitEntry, "--server.port=8080", "--server.address=0.0.0.0")
	} else if wsgi != "" {
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
	return fmt.Sprintf("_startup() { %s; }; ", command)
}

func determineStartCmd(ctx *pythonPlanContext) string {
	serverless := getServerless(ctx)

	// serverless function doesn't need a start command
	if serverless {
		return ""
	}

	startupFunction := determineDefaultStartupFunction(ctx)

	framework := DetermineFramework(ctx)
	if framework == types.PythonFrameworkReflex {
		return startupFunction
	}

	// if "start_command" in `zbpack.json`, or "ZBPACK_START_COMMAND" in env, use it directly
	if value, err := plan.Cast(ctx.Config.Get(plan.ConfigStartCommand), cast.ToStringE).Take(); err == nil {
		return startupFunction + value
	}

	// Call default startup function directly
	return startupFunction + "_startup"
}

// determinePythonVersion Determine Python Version
func determinePythonVersion(ctx *pythonPlanContext) string {
	if pythonVersion, err := plan.Cast(ctx.Config.Get(ConfigPythonVersion), cast.ToStringE).Take(); err == nil {
		return getPython3Version(pythonVersion)
	}

	pm := DeterminePackageManager(ctx)

	switch pm {
	case types.PythonPackageManagerPoetry:
		return determinePythonVersionWithPoetry(ctx)
	case types.PythonPackageManagerPdm:
		return determinePythonVersionWithPdm(ctx)
	case types.PythonPackageManagerRye:
		return determinePythonVersionWithRye(ctx)
	case types.PythonPackageManagerPipenv:
		return determinePythonVersionWithPipenv(ctx)
	default:
		return defaultPython3Version
	}
}

func determinePythonVersionWithPdm(ctx *pythonPlanContext) string {
	src := ctx.Src

	content, err := utils.ReadFileToUTF8(src, "pyproject.toml")
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

	content, err := utils.ReadFileToUTF8(src, "pyproject.toml")
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

func determinePythonVersionWithRye(ctx *pythonPlanContext) string {
	// We read from `.python-version`.
	// The format of `.python-version` is:
	//
	//		[distribution@][version]
	//
	// We extract the version part only.
	src := ctx.Src
	regex := regexp.MustCompile(`(?:.+?@)?([\d.]+)`)

	content, err := utils.ReadFileToUTF8(src, ".python-version")
	if err != nil {
		return defaultPython3Version
	}

	match := regex.FindSubmatch(content)
	if len(match) > 1 {
		return getPython3Version(string(match[1]))
	}

	return defaultPython3Version
}

func determinePythonVersionWithPipenv(ctx *pythonPlanContext) string {
	src := ctx.Src

	content, err := utils.ReadFileToUTF8(src, "Pipfile")
	if err != nil {
		return defaultPython3Version
	}

	compile := regexp.MustCompile(`python_version\s*=\s*"(.*?)"`)
	submatchs := compile.FindStringSubmatch(string(content))
	if len(submatchs) > 1 {
		return submatchs[1]
	}

	return defaultPython3Version
}

func determineBuildCmd(ctx *pythonPlanContext) string {
	commands := ""

	packageManager := DeterminePackageManager(ctx)
	staticInfo := DetermineStaticInfo(ctx)
	framework := DetermineFramework(ctx)

	if postInstallCmd := getPmPostInstallCmd(packageManager); postInstallCmd != "" {
		commands += "RUN " + postInstallCmd + "\n"
	}

	if buildCommand, err := plan.Cast(ctx.Config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		commands += "RUN " + buildCommand + "\n"
	} else {
		if content, err := utils.ReadFileToUTF8(ctx.Src, "package.json"); err == nil {
			if strings.Contains(string(content), "\"build\":") {
				// for example, "build": "vite build"
				commands += "RUN npm install && npm run build\n"
			}
		}

		if framework == types.PythonFrameworkReflex {
			switch packageManager {
			case types.PythonPackageManagerPoetry:
				commands += `RUN poetry run reflex init
RUN poetry run reflex export --frontend-only --no-zip && mv .web/_static/* /srv/ && rm -rf .web`
			default:
				commands += `RUN reflex init
RUN reflex export --frontend-only --no-zip && mv .web/_static/* /srv/ && rm -rf .web`
			}
		}

		if staticInfo.DjangoEnabled() {
			prefix := getPmStartCmdPrefix(packageManager)
			if prefix != "" {
				prefix += " " // ex. poetry run
			}
			// We need to collect static files if we are using Django.
			commands += "RUN " + prefix + "python manage.py collectstatic --noinput\n"
		}

		if determinePlaywright(ctx) {
			commands += "RUN playwright install\n"
		}
	}

	if slices.Contains(determineAptDependencies(ctx), "imagemagick") {
		// credit: https://discord.com/channels/1060209568820494336/1257750217147809903/1258312298674786386
		commands += `ENV IMAGEMAGICK_BINARY=/usr/bin/convert
RUN echo '<?xml version="1.0" encoding="UTF-8"?>' > /etc/ImageMagick-6/policy.xml && \
    echo '<policymap>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="temporary-path" value="/tmp"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="memory" value="2GiB"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="map" value="4GiB"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="width" value="16KP"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="height" value="16KP"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="area" value="128MP"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="disk" value="16GiB"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="thread" value="4"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="throttle" value="0"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="resource" name="time" value="3600"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="path" rights="read|write" pattern="@*"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="coder" rights="read|write" pattern="*"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '  <policy domain="delegate" rights="read|write" pattern="*"/>' >> /etc/ImageMagick-6/policy.xml && \
    echo '</policymap>' >> /etc/ImageMagick-6/policy.xml`
	}

	return strings.TrimSpace(commands)
}

func determineStreamlitEntry(ctx *pythonPlanContext) string {
	src := ctx.Src
	config := ctx.Config
	se := &ctx.StreamlitEntry

	if entry, err := se.Take(); err == nil {
		return entry
	}

	if streamlitEntry := plan.Cast(config.Get(ConfigStreamlitEntry), cast.ToStringE); streamlitEntry.IsSome() {
		*se = optional.Some(streamlitEntry.Unwrap())
		return se.Unwrap()
	}

	for _, file := range []string{"app.py", "main.py", "streamlit_app.py"} {
		content, err := utils.ReadFileToUTF8(src, file)
		if err == nil && bytes.Contains(content, []byte("import streamlit")) {
			*se = optional.Some(file)
			return se.Unwrap()
		}
	}

	*se = optional.Some("")
	return se.Unwrap()
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration
}

func getServerless(ctx *pythonPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
}

func determinePlaywright(ctx *pythonPlanContext) bool {
	return HasDependency(ctx, "playwright")
}

// GetMeta returns the metadata of a Python project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	meta := types.PlanMeta{}

	ctx := &pythonPlanContext{
		Src:    opt.Src,
		Config: opt.Config,
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

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = "true"
	}

	// if serverless, we need a wsgi entry
	if serverless {
		wsgi := DetermineWsgi(ctx)
		if wsgi != "" {
			meta["entry"] = wsgi
		}
	}

	startCmd := determineStartCmd(ctx)
	if startCmd != "" {
		meta["start"] = startCmd
	}

	// if selenium, we need to install chromium
	if HasDependency(ctx, "seleniumbase") || HasDependency(ctx, "selenium") {
		meta["selenium"] = "true"
	}

	aptDeps := determineAptDependencies(ctx)
	if len(aptDeps) > 0 {
		meta["apt-deps"] = strings.Join(aptDeps, " ")
	}

	return meta
}
