package transformer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformStatic copies the static files from output to .zeabur/output/static
// and creates a config.json file for SPA
func TransformStatic(ctx *Context) error {
	if ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	ctx.Log("Transforming static files...\n")

	err := cp.Copy(ctx.BuildkitPath, path.Join(ctx.AppPath, ".zeabur/output/static"))
	if err != nil {
		return fmt.Errorf("copy static files from buildkit output to .zeabur/output/static: %w", err)
	}

	// delete hidden files and directories in output directory
	err = deleteHiddenFilesAndDirs(path.Join(ctx.AppPath, ".zeabur/output/static"))
	if err != nil {
		return fmt.Errorf("delete hidden files and directories in directory: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: make([]types.ZeaburOutputConfigRoute, 0)}
	if isNotMpaFramework(ctx.PlanMeta["framework"]) {
		config.Routes = []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/"}}
	}

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

func isMpaFramework(framework string) bool {
	mpaFrameworks := []types.NodeProjectFramework{
		types.NodeProjectFrameworkHexo,
		types.NodeProjectFrameworkVitepress,
		types.NodeProjectFrameworkAstroStatic,
		types.NodeProjectFrameworkAstroStarlight,
		types.NodeProjectFrameworkSolidStartStatic,
		types.NodeProjectFrameworkVocs,
		types.NodeProjectFrameworkRspress,
	}

	for _, f := range mpaFrameworks {
		if framework == string(f) {
			return true
		}
	}

	return false
}

// isNotMpaFramework is `!isMpaFramework()`, but it's easier to read
func isNotMpaFramework(framework string) bool {
	return !isMpaFramework(framework)
}

// DeleteHiddenFilesAndDirs deletes hidden files and directories in a directory
func deleteHiddenFilesAndDirs(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer func() {
		err := dir.Close()
		if err != nil {
			log.Println("delete hidden files and directories in directory: %w", err)
		}
	}()

	entries, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == ".well-known" {
			continue
		}

		if strings.HasPrefix(entry.Name(), ".") {
			entryPath := filepath.Join(dirPath, entry.Name())

			if entry.IsDir() {
				if err := os.RemoveAll(entryPath); err != nil {
					return err
				}
			} else {
				if err := os.Remove(entryPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
