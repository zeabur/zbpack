package php

import (
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

// DetermineAptDependencies determines the required apt dependencies of the project.
func DetermineAptDependencies(source afero.Fs) []string {
	var dependencies []string

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

	dependenciesUnique := lo.Uniq(dependencies)
	slices.Sort(dependenciesUnique)

	return dependencies
}

// DeterminePHPExtensions determines the required PHP extensions from composer.json of the project.
func DeterminePHPExtensions(source afero.Fs) (extensions []string) {
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

	extensionsUnique := lo.Uniq(extensions)
	slices.Sort(extensionsUnique)

	return extensionsUnique
}

// DetermineStartCommand determines the start command of the project.
func DetermineStartCommand(config plan.ImmutableProjectConfiguration) string {
	completeStartCommand := "_startup() { nginx; php-fpm; }; "

	if startCommand, err := plan.Cast(config.Get(plan.ConfigStartCommand), cast.ToStringE).Take(); err == nil {
		completeStartCommand += startCommand
	} else {
		completeStartCommand += "_startup"
	}

	return completeStartCommand
}

// DetermineBuildCommand determines the build command of the project.
func DetermineBuildCommand(config plan.ImmutableProjectConfiguration) string {
	if buildCommand, err := plan.Cast(config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		return buildCommand
	}

	return ""
}

// DeterminePHPOptimize determines if we should run optimization on build.
func DeterminePHPOptimize(config plan.ImmutableProjectConfiguration) bool {
	return plan.Cast(config.Get(ConfigPHPOptimize), cast.ToBoolE).TakeOr(true)
}
