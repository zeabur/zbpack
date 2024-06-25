package deno

import (
	"encoding/json"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	t "github.com/zeabur/zbpack/pkg/types"
)

// DetermineFramework determines the framework of the Deno project.
func DetermineFramework(src afero.Fs) t.DenoFramework {
	if utils.HasFile(src, "fresh.gen.ts") {
		return t.DenoFrameworkFresh
	}

	return t.DenoFrameworkNone
}

// DetermineEntry determines the entry point of the Deno project.
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

// GetStartCommand gets the start command of the Deno project.
func GetStartCommand(src afero.Fs) string {
	denoJSONMarshal, err := utils.ReadFileToUTF8(src, "deno.json")
	if err != nil {
		return ""
	}

	denoJSON := struct {
		Scripts map[string]string `json:"tasks"`
	}{}

	if err := json.Unmarshal(denoJSONMarshal, &denoJSON); err != nil {
		return ""
	}

	if _, ok := denoJSON.Scripts["start"]; ok {
		return denoJSON.Scripts["start"]
	}

	return ""
}
