package pythonproc

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/prepare", func(_ zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		return &PrepareAction{}, nil
	})
}

// PrepareAction is a procedure that prepares a Python environment.
type PrepareAction struct{}

// Run prepares a Python environment and writes it as a job variable.
func (p PrepareAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	cleanupStack := zbaction.CleanupStack{}
	cleanupFn := cleanupStack.WrapRun()

	venvPath, err := os.MkdirTemp("", "zbpack-python-venv-*")
	if err != nil {
		return cleanupFn, fmt.Errorf("make temp dir: %w", err)
	}
	cleanupStack.Push(func() {
		_ = os.RemoveAll(venvPath)
	})

	// Create virtualenv in this directory.
	cmd := exec.CommandContext(ctx, "uv", "venv", venvPath)
	cmd.Dir = sc.Root()
	cmd.Stdout = sc.Stdout()
	cmd.Stderr = sc.Stderr()
	if err := cmd.Run(); err != nil {
		return cleanupFn, fmt.Errorf("create venv: %w", err)
	}

	jobContext := sc.JobContext()
	venv.RegisterVenvContext(jobContext.ID(), &venv.VirtualEnvironmentContext{
		Path: venvPath,
	})
	cleanupStack.Push(func() {
		venv.DropVenvContext(jobContext.ID())
	})

	return cleanupFn, nil
}
