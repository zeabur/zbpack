package pythonproc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
	"github.com/zeabur/zbpack/pkg/types"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/prepare", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		return &PrepareAction{
			PackageManager: zbaction.NewArgument(args["package-manager"], mapPackageManager),
		}, nil
	})
}

// PrepareAction is a procedure that prepares a Python environment.
type PrepareAction struct {
	PackageManager zbaction.Argument[types.PythonPackageManager]
}

// Run prepares a Python environment and writes it as a job variable.
func (p PrepareAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	cleanupStack := zbaction.CleanupStack{}
	cleanupFn := cleanupStack.WrapRun()

	packageManager := p.PackageManager.Value(sc.ExpandString)

	var venvPath string
	var venvPathGetter func() (string, error)
	var err error

	switch packageManager {
	case types.PythonPackageManagerPip:
		venvPath, err = newVenv(ctx, sc, &cleanupStack)
		if err != nil {
			return cleanupFn, fmt.Errorf("create venv: %w", err)
		}

	case types.PythonPackageManagerPipenv:
		// it is created when running pipenv install,
		// so we don't know at this moment
		//
		// use PathGetter to get the path lazily
		venvPath = ""
		venvPathGetter = func() (string, error) {
			cmd := exec.CommandContext(ctx, "pipenv", "--venv")
			cmd.Dir = sc.Root()

			path, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("get pipenv venv path: %w", err)
			}

			return string(path), nil
		}

	case types.PythonPackageManagerPoetry:
		// it is created when running poetry,
		// so we don't know at this moment
		//
		// use PathGetter to get the path lazily
		venvPath = ""
		venvPathGetter = func() (string, error) {
			cmd := exec.CommandContext(ctx, "poetry", "env", "info", "--path")
			cmd.Dir = sc.Root()

			path, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("get poetry venv path: %w", err)
			}

			return string(path), nil
		}

	case types.PythonPackageManagerPdm, types.PythonPackageManagerRye:
		// it is created when running pdm/rye
		// but its path is fixed
		venvPath = filepath.Join(sc.Root(), ".venv")

	default:
		return cleanupFn, fmt.Errorf("unknown package manager: %s", packageManager)
	}

	jobContext := sc.JobContext()
	venv.RegisterVenvContext(jobContext.ID(), &venv.VirtualEnvironmentContext{
		PackageManager: packageManager,
		Path:           venvPath,
		PathGetter:     venvPathGetter,
	})
	cleanupStack.Push(func() {
		venv.DropVenvContext(jobContext.ID())
	})

	return cleanupFn, nil
}

func newVenv(ctx context.Context, sc *zbaction.StepContext, cleanupStack *zbaction.CleanupStack) (string, error) {
	venvPath, err := os.MkdirTemp("", "zbpack-python-venv-*")
	if err != nil {
		return "", fmt.Errorf("make temp dir: %w", err)
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
		return "", fmt.Errorf("create venv: %w", err)
	}

	return venvPath, nil
}
