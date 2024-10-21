// Package transformer transforms a Docker image TAR output to `.zeabur` directory.
package transformer

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// ErrSkip is a flag for Transformer to skip the transformation,
// which is useful for the transformer that doesn't support such plan.
var ErrSkip = errors.New("skip transformer")

// Context is the context for the transformer.
type Context struct {
	PlanType types.PlanType
	PlanMeta types.PlanMeta

	BuildkitPath afero.Fs
	AppPath      afero.Fs

	PushImage   bool
	ResultImage string
	LogWriter   io.Writer
}

// ZeaburPath returns the `.zeabur` directory of the App path.
func (c *Context) ZeaburPath() afero.Fs {
	if exists, err := afero.DirExists(c.AppPath, ".zeabur"); !exists || err != nil {
		err = c.AppPath.Mkdir(".zeabur", 0o755)
		if err != nil {
			panic(fmt.Errorf("failed to create .zeabur directory: %w", err))
		}
	}

	return afero.NewBasePathFs(c.AppPath, ".zeabur")
}

// Log writes a log message to the log writer.
func (c *Context) Log(format string, args ...interface{}) {
	if c.LogWriter == nil {
		c.LogWriter = os.Stderr
	}

	_, _ = fmt.Fprintf(c.LogWriter, format, args...)
}

// Transformer is the type for the transformer interface.
type Transformer func(ctx *Context) error

// Transform runs the transformers in this package.
func Transform(ctx *Context) error {
	transformers := []Transformer{
		TransformZeaburDir,
		TransformNix,
		TransformGolang,
		TransformRust,
		TransformPython,
	}

	for tid, t := range transformers {
		err := t(ctx)
		switch true {
		case errors.Is(err, ErrSkip):
			continue
		case err != nil:
			return fmt.Errorf("transformer #%d: %w", tid, err)
		case err == nil:
			return nil
		}
	}

	return nil
}
