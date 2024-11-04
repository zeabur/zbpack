package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformRust transforms Rust functions.
func TransformRust(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeRust || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Rust functions...\n")
	zeaburPath := ctx.ZeaburPath()

	err := cp.Copy(ctx.BuildkitPath, filepath.Join(zeaburPath, "output/functions/__rust.func"))
	if err != nil {
		return fmt.Errorf("copy Rust functions: %w", err)
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

	err = funcConfig.WriteTo(filepath.Join(zeaburPath, "output/functions/__rust.func"))
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__rust"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(filepath.Join(zeaburPath, "output/config.json"), configBytes, 0o644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
