package transformer

import (
	"fmt"

	"github.com/spf13/afero"
	"go.nhat.io/aferocopy/v2"
)

// TransformZeaburDir is a transformer function to copy the .zeabur directory.
func TransformZeaburDir(ctx *Context) error {
	if contains, err := afero.DirExists(ctx.BuildkitPath, ".zeabur"); !contains || err != nil {
		return ErrSkip
	}

	ctx.Log("Transforming .zeabur directory...\n")

	err := aferocopy.Copy(".zeabur", ".zeabur", aferocopy.Options{
		SrcFs:  ctx.BuildkitPath,
		DestFs: ctx.AppPath,
	})
	if err != nil {
		return fmt.Errorf("copy .zeabur directory: %w", err)
	}

	return nil
}
