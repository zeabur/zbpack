package transformer

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
	"go.nhat.io/aferocopy/v2"
)

// TransformGolang transforms Golang functions.
func TransformGolang(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeGo || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Golang functions...\n")

	err := aferocopy.Copy("", ".zeabur/output/functions/__go.func", aferocopy.Options{
		SrcFs:  ctx.BuildkitPath,
		DestFs: ctx.AppPath,
	})
	if err != nil {
		return fmt.Errorf("copy Golang functions: %w", err)
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

	err = funcConfig.WriteToFs(ctx.AppPath, ".zeabur/output/functions/__go.func")
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__go"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = afero.WriteFile(ctx.AppPath, ".zeabur/output/config.json", configBytes, 0o644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
