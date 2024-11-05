package transformer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNodejsWaku transforms Node.js Waku framework.
func TransformNodejsWaku(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeNodejs || ctx.PlanMeta["framework"] != "waku" || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	// create a tmpDir to store the build output
	tmpDir, err := os.MkdirTemp("", "zeabur-nodejs-waku-")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// /tmpDir/dist
	wakuDistDir := path.Join(tmpDir, "dist")

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

	err = cp.Copy(path.Join(wakuDistDir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = os.MkdirAll(path.Join(zeaburOutputDir, "functions"), 0o755)
	if err != nil {
		return fmt.Errorf("create functions dir: %w", err)
	}

	_ = os.MkdirAll(path.Join(zeaburOutputDir, "functions/RSC.func"), 0o755)
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

	err = os.WriteFile(path.Join(zeaburOutputDir, "functions/RSC.func/index.mjs"), []byte(indexJS), 0o644)
	if err != nil {
		return fmt.Errorf("write index.js: %w", err)
	}

	config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/RSC"}}}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(ctx.AppPath, ".zeabur/output/config.json"), configBytes, 0o644)
	if err != nil {
		return err
	}

	packageJSON := `{
  "type": "module"
}
`
	err = os.WriteFile(path.Join(zeaburOutputDir, "functions/RSC.func/package.json"), []byte(packageJSON), 0o644)
	if err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}

	return nil
}
