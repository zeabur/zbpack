package deno

import (
	"encoding/json"
	"os"
	"path"
	// "strings"
	"github.com/zeabur/zbpack/internal/utils"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineFramework(absPath string) DenoFramework {
	// Don't ignore fresh.gen.ts
	if utils.HasFile(absPath, "fresh.gen.ts") {
		return DenoFrameworkFresh
	}

	return DenoFrameworkNone
}

func DetermineEntry(absPath string) string {
	//TODO: ts, js, index, main, app.
	if utils.HasFile(absPath, "main.ts") {
		return "main.ts"
	}

	if utils.HasFile(absPath, "main.js") {
		return "main.js"
	}

	return ""
}

func GetStartCommand(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Scripts         map[string]string `json:"scripts"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return ""
	}

	if _, ok := packageJson.DevDependencies["@builder.io/qwik"]; ok {
		if _, ok := packageJson.Scripts["deploy"]; ok {
			return "deploy"
		}
	}

	if _, ok := packageJson.Scripts["start"]; ok {
		return "start"
	}

	return ""
}