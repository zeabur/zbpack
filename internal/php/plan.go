package php

import (
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultPHPVersion is the default PHP version.
const DefaultPHPVersion = "8"

// GetPHPVersion gets the php version of the project.
func GetPHPVersion(source afero.Fs) string {
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

	composerJSON, err := parseComposerJSON(source)
	if err != nil {
		return dependencies
	}

	if composerJSON.Require == nil {
		return dependencies
	}

	// loop through the composer.json dependencies and
	// check if any dependency need some additional apt dependencies
	for dep := range *composerJSON.Require {
		if val, ok := depMap[dep]; ok {
			dependencies = append(dependencies, val...)
		}
	}

	return dependencies
}

var baseExt = []string{
	// php applications often access MySQL databases
	"pdo", "pdo_mysql", "mysqli",
}

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
