package deno

import (
	"encoding/json"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineFramework(src afero.Fs) DenoFramework {
	if utils.HasFile(src, "fresh.gen.ts") {
		return DenoFrameworkFresh
	}

	return DenoFrameworkNone
}

func DetermineEntry(src afero.Fs) string {
	if utils.HasFile(src, "main.ts") {
		return "main.ts"
	}

	if utils.HasFile(src, "main.js") {
		return "main.js"
	}

	if utils.HasFile(src, "index.ts") {
		return "index.ts"
	}

	if utils.HasFile(src, "index.js") {
		return "index.js"
	}

	if utils.HasFile(src, "app.ts") {
		return "app.ts"
	}

	if utils.HasFile(src, "app.js") {
		return "app.js"
	}

	return ""
}

func GetStartCommand(src afero.Fs) string {
	denoJsonMarshal, err := afero.ReadFile(src, "deno.json")
	if err != nil {
		return ""
	}

	denoJson := struct {
		Scripts map[string]string `json:"tasks"`
	}{}

	if err := json.Unmarshal(denoJsonMarshal, &denoJson); err != nil {
		return ""
	}

	if _, ok := denoJson.Scripts["start"]; ok {
		return denoJson.Scripts["start"]
	}

	return ""
}
