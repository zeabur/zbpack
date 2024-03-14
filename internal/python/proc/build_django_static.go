package pythonproc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
	"github.com/zeabur/zbpack/pkg/types"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/build-django-static", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		return &BuildDjangoStaticAction{
			PackageManager: zbaction.NewArgument(args["package-manager"], mapPackageManager),
		}, nil
	})
}

// BuildDjangoStaticAction is a procedure that builds the static files for a Django project.
type BuildDjangoStaticAction struct {
	PackageManager zbaction.Argument[types.PythonPackageManager]
}

// Run builds the static files for a Django project.
func (b BuildDjangoStaticAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	packageManager := b.PackageManager.Value(sc.ExpandString)
	djangoStaticBuildCommand := []string{"python", "manage.py", "collectstatic", "--noinput"}

	// Retrieve a virtual environment.
	jobContext := sc.JobContext()
	venvContext, err := venv.GetVenvContext(jobContext.ID())
	if err != nil {
		return nil, fmt.Errorf("get venv context: %w", err)
	}

	cmdEnv := venvContext.PutEnv(zbaction.ListEnvironmentVariables(sc.VariableContainer()))

	// Run the command.
	{
		exe, args := getRunCommand(packageManager)
		if exe == "" {
			exe = execLookup(djangoStaticBuildCommand[0], cmdEnv["PATH"])
			args = djangoStaticBuildCommand[1:]
		} else {
			args = append(args, djangoStaticBuildCommand...)
		}

		cmd := exec.CommandContext(ctx, exe, args...)
		cmd.Dir = sc.Root()
		cmd.Stdout = sc.Stdout()
		cmd.Stderr = sc.Stderr()
		cmd.Env = cmdEnv.ToList()
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("build Django static files: %w", err)
		}
	}

	return nil, nil
}

func getRunCommand(pm types.PythonPackageManager) (string, []string) {
	switch pm {
	case types.PythonPackageManagerPipenv:
		return "pipenv", []string{"run"}
	case types.PythonPackageManagerPoetry:
		return "poetry", []string{"run"}
	case types.PythonPackageManagerPdm:
		return "pdm", []string{"run"}
	case types.PythonPackageManagerRye:
		return "rye", []string{"run"}
	default:
		return "", nil
	}
}

func execLookup(exe string, pathList string) string {
	for _, path := range strings.Split(pathList, ":") {
		if path == "" {
			continue
		}

		if stat, err := os.Stat(filepath.Join(path, exe)); err == nil {
			// check the file is executable
			if stat.Mode()&0111 != 0 {
				return filepath.Join(path, exe)
			}
		}
	}

	return exe
}
