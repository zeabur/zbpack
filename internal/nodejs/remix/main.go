// Package remix is used to transform build output of Remix app to the serverless build output format of Zeabur
package remix

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	uuid2 "github.com/google/uuid"
	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformServerless will transform the build output of Remix app to the serverless build output format of Zeabur
func TransformServerless(workdir string) error {

	// create a tmpDir to store the build output
	uuid := uuid2.New().String()
	tmpDir := path.Join(os.TempDir(), uuid)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Printf("remove tmp dir: %s\n", err)
		}
	}()

	// /tmpDir/uuid/dist
	remixBuildDir := path.Join(tmpDir, "build")

	// /workDir/.zeabur/output
	zeaburOutputDir := path.Join(workdir, ".zeabur/output")

	fmt.Println("=> Copying build output from image")
	err := cp.Copy(path.Join(os.TempDir(), "zbpack/buildkit"), path.Join(tmpDir))
	if err != nil {
		return err
	}

	fmt.Println("=> Copying static asset files")

	err = os.MkdirAll(path.Join(zeaburOutputDir, "static"), 0755)
	if err != nil {
		return fmt.Errorf("create static dir: %w", err)
	}

	err = cp.Copy(path.Join(tmpDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	fmt.Println("=> Copying remix build output")

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	_ = os.MkdirAll(path.Join(zeaburOutputDir, "functions/index.func"), 0755)

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
		return fmt.Errorf("Failed to write function config to \".zeabur/output/functions/index.func\": " + err.Error())
	}

	fmt.Println("=> Copying node_modules")

	err = cp.Copy(path.Join(remixBuildDir, "../node_modules"), path.Join(zeaburOutputDir, "functions/index.func/node_modules"))
	if err != nil {
		return fmt.Errorf("copy node_modules/waku dir: %w", err)
	}

	fmt.Println("=> Writing config.json ...")

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(workdir, ".zeabur/output/config.json"), configBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Println("=> Copying package.json ...")

	err = cp.Copy(path.Join(tmpDir, "package.json"), path.Join(zeaburOutputDir, "functions/index.func/package.json"))
	if err != nil {
		return fmt.Errorf("copy package.json: %w", err)
	}

	return nil
}
