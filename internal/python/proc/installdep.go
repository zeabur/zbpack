package pythonproc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

// Run install the listed and projects dependencies.
func (a InstallDependenciesAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	packageManager := a.PackageManager.Value(sc.ExpandString)
	extraDependencies := a.ExtraDependencies.Value(sc.ExpandString)

	// Retrieve a virtual environment.
	jobContext := sc.JobContext()
	venvContext, err := venv.GetVenvContext(jobContext.ID())
	if err != nil {
		return nil, fmt.Errorf("get venv context: %w", err)
	}

	// Get requirements.txt content
	requirementContent, err := getRequirementContent(packageManager, sc)
	if err != nil {
		return nil, fmt.Errorf("get requirement content: %w", err)
	}
	slog.Info("requirement", slog.String("content", requirementContent))

	cmdEnv := venvContext.PutEnv(zbaction.ListEnvironmentVariables(sc.VariableContainer())).ToList()

	// Install the extra dependencies first.
	if len(extraDependencies) > 1 {
		exe, args := getAddCommand(extraDependencies...)
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
		exe, args := getInstallCommand()
		cmd := exec.CommandContext(ctx, exe, args...)
		cmd.Dir = sc.Root()
		cmd.Stdin = strings.NewReader(requirementContent)
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

func getRequirementContent(pm types.PythonPackageManager, sc *zbaction.StepContext) (string, error) {
	switch pm {
	case types.PythonPackageManagerPip:
		content, err := os.ReadFile("requirements.txt")
		if err != nil {
			return "", fmt.Errorf("read requirements.txt: %w", err)
		}
		return string(content), nil
	case types.PythonPackageManagerPipenv:
		cmd := exec.Command("pipenv", "requirements")
		cmd.Dir = sc.Root()

		content, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("exec pipenv requirements: %w", err)
		}

		return string(content), nil
	case types.PythonPackageManagerPoetry:
		cmd := exec.Command("poetry", "export", "--without-hashes")
		cmd.Dir = sc.Root()

		content, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("exec poetry export: %w", err)
		}

		return string(content), nil
	case types.PythonPackageManagerPdm:
		cmd := exec.Command("pdm", "export", "--no-hashes", "--no-markers")
		cmd.Dir = sc.Root()

		content, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("exec pdm export: %w", err)
		}

		return string(content), nil
	case types.PythonPackageManagerRye:
		content, err := os.ReadFile("requirements.lock")
		if err != nil {
			return "", fmt.Errorf("read requirements.lock: %w", err)
		}

		return string(content), nil
	}

	return "", fmt.Errorf("unsupported package manager: %s", pm)
}

func getAddCommand(deps ...string) (string, []string) {
	return "uv", append([]string{"pip", "install"}, deps...)
}

func getInstallCommand() (string, []string) {
	return "uv", []string{"pip", "install", "-r", "-"}
}
