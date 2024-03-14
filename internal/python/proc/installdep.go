package pythonproc

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
	"github.com/zeabur/zbpack/pkg/types"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/install-dependencies", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		return &InstallDependenciesAction{
			PackageManager: zbaction.NewArgument(args["package-manager"], mapPackageManager),
			ExtraDependencies: zbaction.NewArgument(args["extra"], func(s string) []string {
				return strings.Split(s, " ")
			}),
		}, nil
	})
}

// InstallDependenciesAction is a procedure that builds a Go binary.
type InstallDependenciesAction struct {
	PackageManager    zbaction.Argument[types.PythonPackageManager]
	ExtraDependencies zbaction.Argument[[]string]
}

// Run installs the listed and projects dependencies.
func (a InstallDependenciesAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	packageManager := a.PackageManager.Value(sc.ExpandString)
	extraDependencies := a.ExtraDependencies.Value(sc.ExpandString)

	// Retrieve a virtual environment.
	jobContext := sc.JobContext()
	venvContext, err := venv.GetVenvContext(jobContext.ID())
	if err != nil {
		return nil, fmt.Errorf("get venv context: %w", err)
	}

	cmdEnv := venvContext.PutEnv(zbaction.ListEnvironmentVariables(sc.VariableContainer())).ToList()

	// Install the extra dependencies first.
	if len(extraDependencies) > 1 {
		exe, args := getAddCommand(packageManager, extraDependencies...)
		cmd := exec.CommandContext(ctx, exe, args...)
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = cmdEnv
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("install extra dependencies: %w", err)
		}
	}

	// Install the project dependencies.
	{
		exe, args := getInstallCommand(packageManager)
		cmd := exec.CommandContext(ctx, exe, args...)
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = cmdEnv
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("install project dependencies: %w", err)
		}
	}

	return nil, nil
}

func mapPackageManager(pm string) types.PythonPackageManager {
	switch pm {
	case "pipenv":
		return types.PythonPackageManagerPipenv
	case "poetry":
		return types.PythonPackageManagerPoetry
	case "pdm":
		return types.PythonPackageManagerPdm
	case "rye":
		return types.PythonPackageManagerRye
	case "pip":
		return types.PythonPackageManagerPip
	default:
		return types.PythonPackageManagerPip
	}
}

func getAddCommand(pm types.PythonPackageManager, deps ...string) (string, []string) {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pipenv", append([]string{"install"}, deps...)
	case types.PythonPackageManagerPoetry:
		return "poetry", append([]string{"add"}, deps...)
	case types.PythonPackageManagerPdm:
		return "pdm", append([]string{"add"}, deps...)
	case types.PythonPackageManagerRye:
		return "rye", append([]string{"add"}, deps...)
	default:
		// our hacked rye uses pip to install dependencies.
		return "uv", append([]string{"pip", "install"}, deps...)
	}
}

func getInstallCommand(pm types.PythonPackageManager) (string, []string) {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pipenv", []string{"install"}
	case types.PythonPackageManagerPoetry:
		return "poetry", []string{"install", "--no-root"}
	case types.PythonPackageManagerPdm:
		return "pdm", []string{"install"}
	case types.PythonPackageManagerRye:
		return "rye", []string{"sync", "--no-dev", "--no-lock"}
	default:
		return "uv", []string{"pip", "install", "-r", "requirements.txt"}
	}
}
