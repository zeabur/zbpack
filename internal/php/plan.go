package php

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	. "github.com/zeabur/zbpack/pkg/types"
)

func GetPhpVersion(absPath string) string {
	composerJsonMarshal, err := os.ReadFile(path.Join(absPath, "composer.json"))
	if err != nil {
		return ""
	}
	composerJson := struct {
		Require map[string]string `json:"require"`
	}{}

	if err := json.Unmarshal(composerJsonMarshal, &composerJson); err != nil {
		return "8.0"
	}
	if composerJson.Require["php"] == "" {
		return "8.0"
	}

	versionRange := composerJson.Require["php"]

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
				// insert error handling here
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
				// insert error handling here
			}
			value -= 0.1

			maxVerString = fmt.Sprintf("%f", value)
			return maxVerString
		}
	}

	return "8.1"
}

func DetermineProjectFramework(absPath string) PhpFramework {
	composerJsonMarshal, err := os.ReadFile(path.Join(absPath, "composer.json"))
	if err != nil {
		return PhpFrameworkNone
	}

	composerJson := struct {
		Name       string            `json:"name"`
		Require    map[string]string `json:"require"`
		Requiredev map[string]string `json:"require-dev"`
	}{}
	if err := json.Unmarshal(composerJsonMarshal, &composerJson); err != nil {
		return PhpFrameworkNone
	}

	if _, isLaravel := composerJson.Require["laravel/framework"]; isLaravel {
		return PhpFrameworkLaravel
	}

	if _, isThinkPHP := composerJson.Require["topthink/framework"]; isThinkPHP {
		return PhpFrameworkThinkphp
	}

	if _, isCodeIgniter := composerJson.Require["codeigniter4/framework"]; isCodeIgniter {
		return PhpFrameworkCodeigniter
	}

	return PhpFrameworkNone

}
