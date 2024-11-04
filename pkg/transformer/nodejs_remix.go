package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNodejsRemix transforms Node.js Remix.js functions.
func TransformNodejsRemix(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeNodejs || ctx.PlanMeta["framework"] != string(types.NodeProjectFrameworkRemix) || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	tmpDir, err := os.MkdirTemp("", "zeabur-nodejs-remix-")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// /tmpDir/dist
	remixBuildDir := path.Join(tmpDir, "build")

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

	err = cp.Copy(path.Join(tmpDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	ctx.Log("=> Copying remix build output\n")

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0o755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	_ = os.MkdirAll(path.Join(zeaburOutputDir, "functions/index.func"), 0o755)

	err = cp.Copy(remixBuildDir, path.Join(zeaburOutputDir, "functions/index.func/build"))
	if err != nil {
		return fmt.Errorf("copy %s to %s: %w", remixBuildDir, path.Join(zeaburOutputDir, "functions/index.func/build"), err)
	}

	entry := "remix-serve ./build/index.js"
	if _, err := os.Stat(path.Join(remixBuildDir, "server", "index.js")); err == nil {
		entry = "remix-serve ./build/server/index.js"
	}
	if _, err := os.Stat(path.Join(remixBuildDir, "server", "index.mjs")); err == nil {
		entry = "remix-serve ./build/server/index.mjs"
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "node20", Entry: entry}
	err = funcConfig.WriteTo(path.Join(zeaburOutputDir, "functions/index.func"))
	if err != nil {
		return fmt.Errorf("Failed to write function config to \".zeabur/output/functions/index.func\": %s", err)
	}

	ctx.Log("=> Copying node_modules\n")

	err = cp.Copy(path.Join(remixBuildDir, "../node_modules"), path.Join(zeaburOutputDir, "functions/index.func/node_modules"))
	if err != nil {
		return fmt.Errorf("copy node_modules/waku dir: %w", err)
	}

	ctx.Log("=> Writing config.json ...\n")

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(ctx.AppPath, ".zeabur/output/config.json"), configBytes, 0o644)
	if err != nil {
		return err
	}

	ctx.Log("=> Copying package.json ...\n")

	err = cp.Copy(path.Join(tmpDir, "package.json"), path.Join(zeaburOutputDir, "functions/index.func/package.json"))
	if err != nil {
		return fmt.Errorf("copy package.json: %w", err)
	}

	return nil
}
