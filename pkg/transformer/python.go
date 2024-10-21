package transformer

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
	"go.nhat.io/aferocopy/v2"
)

// TransformPython transforms Python functions.
func TransformPython(ctx *Context) error {
	const funcPathname = ".zeabur/output/functions/__py.func"

	if ctx.PlanType != types.PlanTypePython || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming Python functions...\n")

	err := aferocopy.Copy("", funcPathname, aferocopy.Options{
		SrcFs:  ctx.BuildkitPath,
		DestFs: ctx.AppPath,
	})
	if err != nil {
		return fmt.Errorf("copy Python functions: %w", err)
	}

	// if there is "static" directory in the output, we will copy it to .zeabur/output/static
	dirStatic, errStatic := afero.DirExists(ctx.BuildkitPath, "static")
	if errStatic == nil && dirStatic {
		err = aferocopy.Copy("static", ".zeabur/output/static", aferocopy.Options{
			SrcFs:  ctx.BuildkitPath,
			DestFs: ctx.AppPath,
		})
		if err != nil {
			return fmt.Errorf("copy static directory: %w", err)
		}
	}

	var venvFs afero.Fs
	dirs, err := afero.ReadDir(ctx.AppPath, funcPathname)
	if err == nil {
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}

			sitePackagesPathname := path.Join("lib", "python"+ctx.PlanMeta["pythonVersion"], "site-packages")

			isLibraryDirectory, err := afero.IsDir(ctx.AppPath, path.Join(funcPathname, dir.Name(), sitePackagesPathname))
			if err != nil || !isLibraryDirectory {
				continue
			}

			venvFs = afero.NewBasePathFs(ctx.AppPath, funcPathname)
		}
	}

	if venvFs != nil {
		outFuncFs := afero.NewBasePathFs(ctx.AppPath, funcPathname)

		_ = outFuncFs.RemoveAll(".site-packages")
		err := aferocopy.Copy("", "site-packages", aferocopy.Options{
			SrcFs:  venvFs,
			DestFs: outFuncFs,
		})
		if err != nil {
			return fmt.Errorf("copy site-packages: %w", err)
		}
	}

	pythonVersionWithoutDot := strings.ReplaceAll(ctx.PlanMeta["pythonVersion"], ".", "")
	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "python" + pythonVersionWithoutDot}
	if ctx.PlanMeta["entry"] != "" {
		funcConfig.Entry = ctx.PlanMeta["entry"]
	}

	err = funcConfig.WriteToFs(ctx.AppPath, funcPathname)
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__py"}}}

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
