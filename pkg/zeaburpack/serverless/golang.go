package zbserverless

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformGoServerless transforms the build output to serverless format.
func TransformGoServerless(imageRootDirectory string, dotZeaburDirectory string, _ types.PlanMeta) error {
	println("Transforming build output to serverless format ...")
	err := cp.Copy(imageRootDirectory, filepath.Join(dotZeaburDirectory, "output/functions/__go.func"))
	if err != nil {
		return fmt.Errorf("copy serverless function: %w", err)
	}

	funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

	err = funcConfig.WriteTo(path.Join(dotZeaburDirectory, "output/functions/__go.func"))
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__go"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(path.Join(dotZeaburDirectory, "output/config.json"), configBytes, 0o644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
