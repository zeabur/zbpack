package pythonproc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
	"github.com/zeabur/zbpack/pkg/types"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/prepare", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		packageManager := args["package-manager"]
		if packageManager == "" {
			return nil, fmt.Errorf("package-manager is not set")
		}

		pythonVersion := args["python-version"]
		if pythonVersion == "" {
			return nil, fmt.Errorf("python-version is not set")
		}

		return &PrepareAction{
			PackageManager: zbaction.NewArgument(packageManager, mapPackageManager),
			PythonVersion:  zbaction.NewArgumentStr(pythonVersion),
		}, nil
	})
}

// PrepareAction is a procedure that prepares a Python environment.
type PrepareAction struct {
	PackageManager zbaction.Argument[types.PythonPackageManager]
	PythonVersion  zbaction.Argument[string]
}

// Run prepares a Python environment and writes it as a job variable.
func (p PrepareAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	cleanupStack := zbaction.CleanupStack{}
	cleanupFn := cleanupStack.WrapRun()

	// Create virtualenv in this directory.
	packageManager := p.PackageManager.Value(sc.ExpandString)
	pythonVersion := p.PythonVersion.Value(sc.ExpandString)

	slog.Info("creating virtual environment",
		slog.String("packageManager", string(packageManager)),
		slog.String("pythonVersion", pythonVersion))

	var venvPath string
	switch packageManager {
	case types.PythonPackageManagerRye:
		// it conflicts with uv – we run `rye sync` instead
		cmd := exec.CommandContext(ctx, "rye", "sync")
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = zbaction.ListEnvironmentVariables(sc.VariableContainer()).ToList()
		if err := cmd.Run(); err != nil {
			return cleanupFn, fmt.Errorf("rye sync: %w", err)
		}

		venvPath = filepath.Join(sc.Root(), ".venv")
	default:
		var err error

		venvPath, err = os.MkdirTemp("", "zbpack-python-venv-*")
		if err != nil {
			return cleanupFn, fmt.Errorf("make temp dir: %w", err)
		}
		cleanupStack.Push(func() {
			_ = os.RemoveAll(venvPath)
		})

		cmd := exec.CommandContext(ctx, "uv", "venv", "-p", pythonVersion, venvPath)
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = zbaction.ListEnvironmentVariables(sc.VariableContainer()).ToList()
		if err := cmd.Run(); err != nil {
			return cleanupFn, fmt.Errorf("create venv: %w", err)
		}
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
