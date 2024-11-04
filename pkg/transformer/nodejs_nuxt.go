package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNodejsNuxt transforms Node.js Nuxt.js framework.
func TransformNodejsNuxt(ctx *Context) error {
	if (ctx.PlanType != types.PlanTypeNodejs && ctx.PlanType != types.PlanTypeBun) || !types.IsNitroBasedFramework(ctx.PlanMeta["framework"]) || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	tmpDir, err := os.MkdirTemp("", "zeabur-nodejs-nuxt-")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// /tmpDir/.output
	nuxtOutputDir := path.Join(tmpDir, ".output")

	// /workDir/.zeabur/output
	zeaburOutputDir := path.Join(ctx.AppPath, ".zeabur/output")

	ctx.Log("=> Copying build output from image\n")

	err = cp.Copy(ctx.BuildkitPath, tmpDir)
	if err != nil {
		return err
	}

	ctx.Log("=> Copying static asset files\n")

	err = os.MkdirAll(path.Join(zeaburOutputDir, "static"), 0o755)
	if err != nil {
		return fmt.Errorf("create static dir: %w", err)
	}

	err = cp.Copy(path.Join(nuxtOutputDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0o755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	err = cp.Copy(path.Join(nuxtOutputDir, "server"), path.Join(zeaburOutputDir, "functions/__nitro.func"))
	if err != nil {
		return fmt.Errorf("copy nitro function dir: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__nitro"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(ctx.AppPath, ".zeabur/output/config.json"), configBytes, 0o644)
	if err != nil {
		return err
	}

	return nil
}
