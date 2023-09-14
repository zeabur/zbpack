package static

import (
	"encoding/json"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
	"os"
	"path"
)

// TransformServerless copies the static files from output to .zeabur/output/static and creates a config.json file for SPA
func TransformServerless(image, workdir string, meta types.PlanMeta) error {
	err := utils.CopyFromImage(image, path.Join("/src", meta["outputDir"])+"/.", path.Join(workdir, ".zeabur/output/static"))
	if err != nil {
		return err
	}

	config := types.ZeaburOutputConfig{Containerized: false, Routes: make([]types.ZeaburOutputConfigRoute, 0)}
	if isNotMpaFramework(meta["framework"]) {
		config.Routes = []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/index.html"}}
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(workdir, ".zeabur/output/config.json"), configBytes, 0644)
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
		types.NodeProjectFrameworkSolidStartStatic,
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
