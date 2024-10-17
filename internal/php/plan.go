package php

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultPHPVersion is the default PHP version.
const DefaultPHPVersion = "8"

// GetPHPVersion gets the php version of the project.
func GetPHPVersion(config plan.ImmutableProjectConfiguration, source afero.Fs) string {
	// Priority: config (environment variable) > docker-compose.yml > composer.json

	// Get the PHP version from the config (php.version) or environment variable (ZBPACK_PHP_VERSION).
	if phpVersion, err := plan.Cast(config.Get(ConfigPHPVersion), cast.ToStringE).Take(); err == nil {
		return phpVersion
	}

	// if not found in the config or environment variable, try to get it from the docker-compose.yml because it may be a Laravel Sail project.
	compose, err := utils.ReadFileToUTF8(source, "docker-compose.yml")
	if err == nil && strings.Contains(string(compose), "vendor/laravel/sail/runtimes") {
		lines := strings.Split(string(compose), "\n")
		for _, line := range lines {
			if strings.Contains(line, "vendor/laravel/sail/runtimes") {
				parts := strings.Split(line, "/")
				return parts[len(parts)-1]
			}
		}
	}

	// if not found in the docker-compose.yml, try to get it from the "require.php" of composer.json.

	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return DefaultPHPVersion
	}

	versionRange, ok := composerJSON.GetRequire("php")
	if !ok || versionRange == "" {
		return DefaultPHPVersion
	}

	return utils.ConstraintToVersion(versionRange, DefaultPHPVersion)
}

// DetermineProjectFramework determines the framework of the project.
func DetermineProjectFramework(source afero.Fs) types.PHPFramework {
	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return types.PHPFrameworkNone
	}

	if _, isLaravel := composerJSON.GetRequire("laravel/framework"); isLaravel {
		return types.PHPFrameworkLaravel
	}

	if _, isThinkPHP := composerJSON.GetRequire("topthink/framework"); isThinkPHP {
		return types.PHPFrameworkThinkphp
	}

	if _, isCodeIgniter := composerJSON.GetRequire("codeigniter4/framework"); isCodeIgniter {
		return types.PHPFrameworkCodeigniter
	}

	if _, isSymfony := composerJSON.GetRequire("symfony/runtime"); isSymfony {
		return types.PHPFrameworkSymfony
	}

	return types.PHPFrameworkNone
}

var depMap = map[string][]string{
	"ext-openssl": {"libssl-dev"},
	"ext-zip":     {"libzip-dev"},
	"ext-curl":    {"libcurl4-openssl-dev", "libssl-dev"},
	"ext-gd":      {"libpng-dev"},
	"ext-gmp":     {"libgmp-dev"},
}

var baseDep = []string{"libicu-dev", "pkg-config", "unzip", "git"}

// DetermineAptDependencies determines the required apt dependencies of the project.
//
// We install Nginx server unless server is "swoole".
func DetermineAptDependencies(source afero.Fs, server string) []string {
	// deep copy the base dependencies
	dependencies := slices.Clone(baseDep)

	// If Octane Server is not "swoole", we should install Nginx.
	//
	// TODO: support RoadRunner
	if server != "swoole" {
		dependencies = append(dependencies, "nginx")
	}

	// Install Node.js if package.json exists.
	if exists, _ := afero.Exists(source, "package.json"); exists {
		dependencies = append(dependencies, "nodejs", "npm")
	}

	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return dependencies
	}

	if composerJSON.Require == nil {
		return dependencies
	}

	// loop through the composer.json dependencies and
	// check if any dependency needs some additional apt dependencies
	for dep := range *composerJSON.Require {
		if val, ok := depMap[dep]; ok {
			dependencies = append(dependencies, val...)
		}
	}

	return dependencies
}

var baseExt = []string{"pdo", "pdo_mysql", "mysqli", "gd", "curl", "zip", "intl", "pcntl", "bcmath"}

// DeterminePHPExtensions determines the required PHP extensions from composer.json of the project.
func DeterminePHPExtensions(source afero.Fs) []string {
	extensions := slices.Clone(baseExt)

	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return extensions
	}

	if composerJSON.Require == nil {
		return extensions
	}

	for dep := range *composerJSON.Require {
		extName, ok := strings.CutPrefix(dep, "ext-")
		if ok {
			extensions = append(extensions, strings.ToLower(extName))
		}
	}

	return lo.Uniq(extensions)
}

// DetermineApplication determines what application the project is using.
// Therefore, we can apply some custom fixes such as the nginx configuration.
func DetermineApplication(source afero.Fs) (types.PHPApplication, types.PHPProperty) {
	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return types.PHPApplicationDefault, types.PHPPropertyNone
	}

	if composerJSON.Name == "lizhipay/acg-faka" {
		return types.PHPApplicationAcgFaka, types.PHPPropertyComposer
	}

	return types.PHPApplicationDefault, types.PHPPropertyComposer
}

// determineStartupFunction determines the startup function of the project.
func determineStartupFunction(config plan.ImmutableProjectConfiguration) string {
	var startupFnBody string

	octaneServerType := plan.Cast(config.Get(ConfigLaravelOctaneServer), castOctaneServer).TakeOr("")
	switch octaneServerType {
	case OctaneServerSwoole:
		startupFnBody = "php artisan octane:start --server=swoole --host=0.0.0.0 --port=8080"
	case OctaneServerRoadrunner: // unimplemented
		fallthrough
	default: // none
		startupFnBody = "nginx; php-fpm"
	}

	return "_startup(){ " + startupFnBody + "; }; "
}

// DetermineStartCommand determines the start command of the project.
func DetermineStartCommand(config plan.ImmutableProjectConfiguration) string {
	completeStartCommand := determineStartupFunction(config)

	if startCommand, err := plan.Cast(config.Get(plan.ConfigStartCommand), cast.ToStringE).Take(); err == nil {
		completeStartCommand += startCommand
	} else {
		completeStartCommand += "_startup"
	}

	return completeStartCommand
}

// DetermineBuildCommand determines the build command of the project.
func DetermineBuildCommand(source afero.Fs, config plan.ImmutableProjectConfiguration) string {
	if buildCommand, err := plan.Cast(config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		return buildCommand
	}
	if content, err := afero.ReadFile(source, "package.json"); err == nil {
		if bytes.Contains(content, []byte("\"build\":")) {
			// "build": "vite build"
			return "npm install && npm run build"
		}
	}

	return ""
}

const (
	// OctaneServerRoadrunner indicates this Laravel Octane server uses RoadRunner.
	OctaneServerRoadrunner = "roadrunner"
	// OctaneServerSwoole indicates this Laravel Octane server uses Swoole.
	OctaneServerSwoole = "swoole"
)

func castOctaneServer(i interface{}) (string, error) {
	s, err := cast.ToStringE(i)
	if err != nil {
		return "", err
	}

	switch s {
	case OctaneServerRoadrunner, OctaneServerSwoole:
		return s, nil
	default:
		return "", fmt.Errorf("unknown octane server: %s", s)
	}
}
