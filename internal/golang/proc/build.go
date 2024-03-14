package golangproc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	zbaction "github.com/zeabur/action"
)

func init() {
	zbaction.RegisterProcedure("zbpack/golang/build", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		entry, ok := args["entry"]
		if !ok {
			entry = "."
		}

		return &BuildAction{
			Entry: zbaction.NewArgumentStr(entry),
		}, nil
	})
}

// BuildAction is a procedure that builds a Go binary.
type BuildAction struct {
	Entry zbaction.Argument[string]
}

// Run builds a Go binary.
func (b BuildAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	entry := b.Entry.Value(sc.ExpandString)

	// Make a directory for storing the binaries.
	outDir, err := os.MkdirTemp("", "zbpack-go-out-*")
	if err != nil {
		return nil, fmt.Errorf("make temp dir: %w", err)
	}

	env := sc.VariableContainer()
	if _, ok := env.GetRawVariable("CGO_ENABLED"); !ok {
		env = zbaction.NewVariableContainerWithExtraParameters(map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
		}, env)
	}

	outFile := path.Join(outDir, "server")
	// Build the binary.
	{
		cmd := exec.CommandContext(ctx, "go", "build", "-o", outFile, entry)
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = zbaction.ListEnvironmentVariables(env).ToList()
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("build: %w", err)
		}
	}

	// Set the output
	sc.SetThisOutput("outDir", outDir)
	sc.SetThisOutput("outFile", outFile)

	// Clean up
	return func() {
		_ = os.RemoveAll(outDir)
	}, nil
}
