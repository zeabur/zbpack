package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformPython transforms Python functions.
func TransformPython(ctx *Context) error {
	const funcPathname = "output/functions/__py.func"

	if ctx.PlanType != types.PlanTypePython || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	zeaburPath := ctx.ZeaburPath()

	ctx.Log("Transforming Python functions...\n")
	err := cp.Copy(ctx.BuildkitPath, filepath.Join(zeaburPath, funcPathname))
	if err != nil {
		return fmt.Errorf("copy Python functions: %w", err)
	}

	// if there is "static" directory in the output, we will copy it to .zeabur/output/static
	statStatic, err := os.Stat(filepath.Join(ctx.BuildkitPath, "static"))
	if err == nil && statStatic.IsDir() {
		err = cp.Copy(filepath.Join(ctx.BuildkitPath, "static"), filepath.Join(zeaburPath, "output/static"))
		if err != nil {
			return fmt.Errorf("copy static directory: %w", err)
		}
	}

	var venvDir string
	dirs, err := os.ReadDir(filepath.Join(zeaburPath, funcPathname))
	if err == nil {
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}

			sitePackagesPathname := path.Join("lib", "python"+ctx.PlanMeta["pythonVersion"], "site-packages")

			statLib, err := os.Stat(filepath.Join(zeaburPath, funcPathname, dir.Name(), sitePackagesPathname))
			if err != nil || !statLib.IsDir() {
				continue
			}

			venvDir = filepath.Join(zeaburPath, funcPathname, dir.Name())
		}
	}

	if venvDir != "" {
		outFuncDir := filepath.Join(zeaburPath, funcPathname)

		_ = os.RemoveAll(filepath.Join(outFuncDir, ".site-packages"))

		err = cp.Copy(venvDir, filepath.Join(outFuncDir, ".site-packages"))
		if err != nil {
			return fmt.Errorf("copy site-packages: %w", err)
		}
	}

	pythonVersionWithoutDot := strings.ReplaceAll(ctx.PlanMeta["pythonVersion"], ".", "")
	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "python" + pythonVersionWithoutDot}
	if ctx.PlanMeta["entry"] != "" {
		funcConfig.Entry = ctx.PlanMeta["entry"]
	}

	err = funcConfig.WriteTo(filepath.Join(zeaburPath, funcPathname))
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__py"}}}

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
