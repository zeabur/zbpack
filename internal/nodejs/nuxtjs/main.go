// Package nuxtjs is used to transform build output of Nuxt.js app to the serverless build output format of Zeabur
package nuxtjs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	uuid2 "github.com/google/uuid"
	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformServerless will transform build output of Nuxt.js app to the serverless build output format of Zeabur
func TransformServerless(workdir string) error {

	// create a tmpDir to store the build output of Next.js app
	uuid := uuid2.New().String()
	tmpDir := path.Join(os.TempDir(), uuid)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Printf("remove tmp dir: %s\n", err)
		}
	}()

	// /tmpDir/uuid/.output
	nuxtOutputDir := path.Join(tmpDir, ".output")

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

	err = cp.Copy(path.Join(nuxtOutputDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0755)
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

	err = os.WriteFile(path.Join(workdir, ".zeabur/output/config.json"), configBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
