package transformer

import (
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

// TransformZeaburDir is a transformer function to copy the .zeabur directory.
func TransformZeaburDir(ctx *Context) error {
	if statZeabur, err := os.Stat(
		filepath.Join(ctx.BuildkitPath, ".zeabur"),
	); err != nil || !statZeabur.IsDir() {
		return ErrSkip
	}

	ctx.Log("Transforming .zeabur directory...\n")

	err := cp.Copy(filepath.Join(ctx.BuildkitPath, ".zeabur"), ctx.ZeaburPath())
	if err != nil {
		return fmt.Errorf("copy .zeabur directory: %w", err)
	}

	return nil
}
