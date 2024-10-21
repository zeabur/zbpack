package transformer

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
	"go.nhat.io/aferocopy/v2"
)

// TransformRust transforms Rust functions.
func TransformRust(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeRust || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Rust functions...\n")

	err := aferocopy.Copy("", ".zeabur/output/functions/__rust.func", aferocopy.Options{
		SrcFs:  ctx.BuildkitPath,
		DestFs: ctx.AppPath,
	})
	if err != nil {
		return fmt.Errorf("copy Rust functions: %w", err)
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

	err = funcConfig.WriteToFs(ctx.AppPath, ".zeabur/output/functions/__rust.func")
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__rust"}}}

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
