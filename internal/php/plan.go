package php

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
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
	versionRange := strings.Split(composerJson.Require["php"], "|")
	if len(versionRange) == 1 {
		fmt.Println("isVersion")
		return strings.Trim(versionRange[0], "^")
	} else {
		fmt.Println("isVersionRange")
		return strings.Trim(versionRange[1], "^")
	}

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
		fmt.Println("isLaravel")
		return PhpFrameworkLaravel
	}

	return PhpFrameworkNone

}
