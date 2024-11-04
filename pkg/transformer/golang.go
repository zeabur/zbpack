package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformGolang transforms Golang functions.
func TransformGolang(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeGo || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Golang functions...\n")
	zeaburPath := ctx.ZeaburPath()

	err := cp.Copy(ctx.BuildkitPath, filepath.Join(zeaburPath, "output/functions/__go.func"))
	if err != nil {
		return fmt.Errorf("copy Golang functions: %w", err)
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

	err = funcConfig.WriteTo(filepath.Join(zeaburPath, "output/functions/__go.func"))
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__go"}}}

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
