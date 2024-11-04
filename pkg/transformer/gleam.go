package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformGleam transforms the Gleam build output to serverless format.
func TransformGleam(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeGleam || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Gleam build output to serverless format ...\n")

	funcPath := filepath.Join(ctx.AppPath, ".zeabur/output/functions/__erl.func")
	err := cp.Copy(ctx.BuildkitPath, funcPath)
	if err != nil {
		return fmt.Errorf("copy serverless function: %w", err)
	}

	content, err := os.ReadFile(filepath.Join(ctx.AppPath, ".zeabur/output/functions/__erl.func/entrypoint.sh"))
	if err != nil {
		return fmt.Errorf("read entrypoint.sh: %w", err)
	}

	entry := utils.ExtractErlangEntryFromGleamEntrypointShell(string(content))
	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "erlang27", Entry: entry}

	err = funcConfig.WriteTo(filepath.Join(ctx.AppPath, ".zeabur/output/functions/__erl.func"))
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__erl"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(ctx.AppPath, ".zeabur/output/config.json"), configBytes, 0o644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
