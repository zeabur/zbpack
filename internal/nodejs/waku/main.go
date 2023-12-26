// Package waku is used to transform build output of Waku app to the serverless build output format of Zeabur
package waku

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

// TransformServerless will transform the build output of Waku app to the serverless build output format of Zeabur
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
	wakuDistDir := path.Join(tmpDir, "dist")

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

	err = cp.Copy(path.Join(wakuDistDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	_ = os.MkdirAll(path.Join(zeaburOutputDir, "functions/RSC.func"), 0755)
	err = cp.Copy(wakuDistDir, path.Join(zeaburOutputDir, "functions/RSC.func/dist"))
	if err != nil {
		return fmt.Errorf("copy waku's RSC function dir: %w", err)
	}

	err = utils.Copy(path.Join(wakuDistDir, "../node_modules/waku"), path.Join(zeaburOutputDir, "functions/RSC.func/node_modules/waku"))
	if err != nil {
		return fmt.Errorf("copy node_modules/waku dir: %w", err)
	}

	indexJS := `import path from 'node:path';
import { connectMiddleware } from 'waku';
const entries = import(path.resolve('dist', 'entries.js'));
export default async function handler(req, res) {
  connectMiddleware({ entries, ssr: true })(req, res, () => {
    res.statusCode = 404;
    res.end();
  });
}
`

	err = os.WriteFile(path.Join(zeaburOutputDir, "functions/RSC.func/index.mjs"), []byte(indexJS), 0644)
	if err != nil {
		return fmt.Errorf("write index.js: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/RSC"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(workdir, ".zeabur/output/config.json"), configBytes, 0644)
	if err != nil {
		return err
	}

	packageJSON := `{
  "type": "module"
}
`
	err = os.WriteFile(path.Join(zeaburOutputDir, "functions/RSC.func/package.json"), []byte(packageJSON), 0644)
	if err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}

	return nil
}
