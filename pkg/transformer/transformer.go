// Package transformer transforms a Docker image TAR output to `.zeabur` directory.
package transformer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeabur/zbpack/pkg/types"
)

// ErrSkip is a flag for Transformer to skip the transformation,
// which is useful for the transformer that doesn't support such plan.
var ErrSkip = errors.New("skip transformer")

// Context is the context for the transformer.
type Context struct {
	PlanType types.PlanType
	PlanMeta types.PlanMeta

	BuildkitPath string
	AppPath      string

	PushImage   bool
	ResultImage string
	LogWriter   io.Writer
}

// ZeaburPath returns the `.zeabur` directory of the App path.
func (c *Context) ZeaburPath() string {
	if _, err := os.Stat(filepath.Join(c.AppPath, ".zeabur")); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(filepath.Join(c.AppPath, ".zeabur"), 0o755)
			if err != nil {
				panic(fmt.Errorf("failed to create .zeabur directory: %w", err))
			}
		} else {
			panic(fmt.Errorf("failed to stat .zeabur directory: %w", err))
		}
	}

	return filepath.Join(c.AppPath, ".zeabur")
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
		TransformNodejsWaku,
		TransformNodejsNext,
		TransformNodejsRemix,
		TransformNodejsNuxt,
		TransformGleam,
		TransformStatic,
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
