// Package umi is used to transform build output of Umi.js app to the serverless build output format of Zeabur
package umi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	uuid2 "github.com/google/uuid"
	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformServerless will transform the build output of Umi.js app to the serverless build output format of Zeabur
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
	umiDistDir := path.Join(tmpDir, "dist")

	// /tmpDir/uuid/api
	umiApiDir := path.Join(tmpDir, "api")

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

	err = cp.Copy(umiDistDir, path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	_ = os.MkdirAll(path.Join(zeaburOutputDir, "functions/__umi.func"), 0755)
	err = cp.Copy(umiApiDir, path.Join(zeaburOutputDir, "functions/__umi.func/api"))
	if err != nil {
		return fmt.Errorf("copy to functions dir: %w", err)
	}

	err = utils.Copy(path.Join(tmpDir, "/node_modules"), path.Join(zeaburOutputDir, "functions/__umi.func/node_modules"))
	if err != nil {
		return fmt.Errorf("copy node_modules dir: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{
		// redirect all requests not match any static files to __umi function
		{Src: "/(.*)", Dest: "/__umi"}}}

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
