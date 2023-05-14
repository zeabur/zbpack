package php

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// GetPHPVersion gets the php version of the project.
func GetPHPVersion(source afero.Fs) string {
	composerJSONMarshal, err := afero.ReadFile(source, "composer.json")
	if err != nil {
		return ""
	}
	composerJSON := struct {
		Require map[string]string `json:"require"`
	}{}

	if err := json.Unmarshal(composerJSONMarshal, &composerJSON); err != nil {
		return "8.0"
	}
	if composerJSON.Require["php"] == "" {
		return "8.0"
	}

	versionRange := composerJSON.Require["php"]

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

	return "8.1"
}

// DetermineProjectFramework determines the framework of the project.
func DetermineProjectFramework(source afero.Fs) types.PHPFramework {
	composerJSONMarshal, err := afero.ReadFile(source, "composer.json")
	if err != nil {
		return types.PHPFrameworkNone
	}

	composerJSON := struct {
		Name       string            `json:"name"`
		Require    map[string]string `json:"require"`
		Requiredev map[string]string `json:"require-dev"`
	}{}
	if err := json.Unmarshal(composerJSONMarshal, &composerJSON); err != nil {
		return types.PHPFrameworkNone
	}

	if _, isLaravel := composerJSON.Require["laravel/framework"]; isLaravel {
		return types.PHPFrameworkLaravel
	}

	if _, isThinkPHP := composerJSON.Require["topthink/framework"]; isThinkPHP {
		return types.PHPFrameworkThinkphp
	}

	if _, isCodeIgniter := composerJSON.Require["codeigniter4/framework"]; isCodeIgniter {
		return types.PHPFrameworkCodeigniter
	}

	return types.PHPFrameworkNone
}
