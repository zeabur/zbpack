package deno

import (
	"encoding/json"
	"os"
	"path"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineFramework(absPath string) DenoFramework {
	if utils.HasFile(absPath, "fresh.gen.ts") {
		return DenoFrameworkFresh
	}

	return DenoFrameworkNone
}

func DetermineEntry(absPath string) string {
	if utils.HasFile(absPath, "main.ts") {
		return "main.ts"
	}

	if utils.HasFile(absPath, "main.js") {
		return "main.js"
	}

	if utils.HasFile(absPath, "index.ts") {
		return "index.ts"
	}

	if utils.HasFile(absPath, "index.js") {
		return "index.js"
	}

	if utils.HasFile(absPath, "app.ts") {
		return "app.ts"
	}

	if utils.HasFile(absPath, "app.js") {
		return "app.js"
	}

	return ""
}

func GetStartCommand(absPath string) string {
	denoJsonMarshal, err := os.ReadFile(path.Join(absPath, "deno.json"))
	if err != nil {
		return ""
	}

	denoJson := struct {
		Scripts         map[string]string `json:"tasks"`
	}{}

	if err := json.Unmarshal(denoJsonMarshal, &denoJson); err != nil {
		return ""
	}

	if _, ok := denoJson.Scripts["start"]; ok {
		return denoJson.Scripts["start"]
	}

	return ""
}