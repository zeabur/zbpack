package php

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// DefaultPHPVersion represents the default PHP version.
const DefaultPHPVersion = "8.1"

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

	isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	if isVersion {
		return versionRange
	}
	ranges := strings.Split(versionRange, " ")
	for _, r := range ranges {
		if strings.HasPrefix(r, ">=") {
			minVerString := strings.TrimPrefix(r, ">=")
			return minVerString
		} else if strings.HasPrefix(r, ">") {
			minVerString := strings.TrimPrefix(r, ">")
			value, err := strconv.ParseFloat(minVerString, 64)
			if err != nil {
				log.Println("parse php version error", err)
				continue
			}
			value += 0.1
			minVerString = fmt.Sprintf("%f", value)
			return minVerString
		} else if strings.HasPrefix(r, "<=") {
			maxVerString := strings.TrimPrefix(r, "<=")
			return maxVerString

		} else if strings.HasPrefix(r, "<") {
			maxVerString := strings.TrimPrefix(r, "<=")
			value, err := strconv.ParseFloat(maxVerString, 64)
			if err != nil {
				log.Println("parse php version error", err)
				continue
			}
			value -= 0.1

			maxVerString = fmt.Sprintf("%f", value)
			return maxVerString
		}
	}

	return DefaultPHPVersion
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
