package pythonproc

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	cp "github.com/otiai10/copy"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/python/venv"
	"github.com/zeabur/zbpack/pkg/types"
)

func init() {
	zbaction.RegisterProcedure("zbpack/python/bundle-docker", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		entrypoint, ok := args["entrypoint"]
		if !ok {
			return nil, zbaction.NewErrRequiredArgument("entrypoint")
		}

		entrypointType, ok := args["entrypoint-type"]
		if !ok {
			return nil, zbaction.NewErrRequiredArgument("entrypoint-type")
		}

		ctx, ok := args["context"]
		if !ok {
			return nil, zbaction.NewErrRequiredArgument("context")
		}

		cache, ok := args["cache"]
		if !ok {
			cache = "true"
		}

		return &BundleDockerAction{
			RuntimeDependencies: zbaction.NewArgument(args["runtime-dependencies"], strings.Fields),
			Entrypoint:          zbaction.NewArgumentStr(entrypoint),
			EntrypointType: zbaction.NewArgument(entrypointType, func(t string) types.PythonEntrypointType {
				return types.PythonEntrypointType(t)
			}),
			StaticDirMap: zbaction.NewArgument(args["static-dir-map"], func(s string) *StaticDirMap {
				hostDir, urlPath, ok := strings.Cut(s, ":")
				if !ok {
					return nil
				}

				return &StaticDirMap{
					StaticHostDir: hostDir,
					StaticURLPath: urlPath,
				}
			}),
			Context: zbaction.NewArgumentStr(ctx),
			Cache:   zbaction.NewArgumentBool(cache),
		}, nil
	})
}

// BundleDockerAction is a procedure that bundles the current project as a Docker image.
type BundleDockerAction struct {
	// RuntimeDependencies is the list of runtime packages to install.
	RuntimeDependencies zbaction.Argument[[]string]

	// Entrypoint is the entrypoint of the project.
	// Usually it is the path to the main script (main.py).
	// For WSGI project, it should be the WSGI path (main:app).
	Entrypoint zbaction.Argument[string]
	// EntrypointType is the type of the entrypoint.
	// For files like "main.py", it should be "file",
	// for WSGI projects, it should be "wsgi".
	EntrypointType zbaction.Argument[types.PythonEntrypointType]

	// StaticDirMap maps the static directory to the URL path for Nginx to serve.
	//
	// For example, if the static directory is in "/app/static", and the URL path is "/static",
	// then you should pass "static-dir-map" as: "/app/static:/static".
	//
	// If no static directory is provided, we will not start a Nginx server to serve the static files.
	StaticDirMap zbaction.Argument[*StaticDirMap]

	// Context is the directory to run the build in.
	Context zbaction.Argument[string]
	// Cache indicates whether to use cache when building the image.
	// By default, it is true.
	Cache zbaction.Argument[bool]
}

// StaticDirMap is a structure saves the relationship of a static directory to a URL path.
type StaticDirMap struct {
	// StaticHostDir indicates where the static files are located in the host.
	// For example, `/app/static` means that the static files are located at
	// `/app/static` in the container.
	StaticHostDir string

	// StaticURLPath indicates where the static files can be accessed from the web.
	// For example, `/static` means the static files can be accessed from `http://example.com/static`.
	StaticURLPath string
}

// Run bundles the current project as a Docker image.
func (a BundleDockerAction) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	// Retrieve a virtual environment.
	jobContext := sc.JobContext()
	venvContext, err := venv.GetVenvContext(jobContext.ID())
	if err != nil {
		return nil, fmt.Errorf("get venv context: %w", err)
	}

	// Get the site_packages of the virtual environment – we will copy it to the Docker image.
	sitePackages, err := venvContext.GetSitePackagesDirectory()
	if err != nil {
		return nil, fmt.Errorf("get site-packages: %w", err)
	}

	pythonLibPath := filepath.Join(sitePackages, "..")
	pythonLibName := filepath.Base(pythonLibPath)

	pythonVersion, ok := strings.CutPrefix(pythonLibName, "python")
	if !ok {
		return nil, fmt.Errorf("not a valid python lib path: %s", pythonLibPath)
	}

	// Get the listening port of the WSGI server.
	staticDirMap := a.StaticDirMap.Value(sc.ExpandString)
	listeningPort := "8080"

	if staticDirMap != nil {
		listeningPort = "8000" // 8080 -> nginx
	}

	// Get the entrypoint of the project.
	entrypoint := a.Entrypoint.Value(sc.ExpandString)
	dockerCmd := "CMD "

	if staticDirMap != nil {
		dockerCmd += "nginx && "
	}

	switch a.EntrypointType.Value(sc.ExpandString) {
	case types.PythonEntrypointTypeFile:
		dockerCmd += fmt.Sprintf("python %s", entrypoint)
	case types.PythonEntrypointTypeWsgi:
		dockerCmd += fmt.Sprintf("gunicorn --bind :%s %s", listeningPort, entrypoint)
	case types.PythonEntrypointTypeAsgi:
		dockerCmd += fmt.Sprintf("uvicorn --host 0.0.0.0 --port %s %s", listeningPort, entrypoint)
	case types.PythonEntrypointTypeStreamlit:
		dockerCmd += fmt.Sprintf("streamlit run %s --server.address=0.0.0.0 --server.port=%s", entrypoint, listeningPort)
	case types.PythonEntrypointTypeSanic:
		dockerCmd += fmt.Sprintf("sanic %s --host 0.0.0.0 --port :%s", entrypoint, listeningPort)
	}

	// Wrap context directory.
	wrappedContextDirectory, err := os.MkdirTemp("", "zbpack-docker-context-*")
	if err != nil {
		return nil, fmt.Errorf("create context directory: %w", err)
	}
	slog.Info("Context directory created", slog.String("dir", wrappedContextDirectory))
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(wrappedContextDirectory)

	// Copy the site-packages to the context directory.
	_ = os.WriteFile(filepath.Join(sitePackages, "COPIED_FROM_LOCAL"), []byte("COPIED_FROM_LOCAL"), 0644)

	containerLibPath := filepath.Join(wrappedContextDirectory, "usr", "local", "lib")
	if err := cp.Copy(pythonLibPath, path.Join(containerLibPath, pythonLibName)); err != nil {
		return nil, fmt.Errorf("copy site-packages to wrapped context directory: %w", err)
	}

	// Copy the bin files to /usr/local/bin
	containerBinPath := filepath.Join(wrappedContextDirectory, "usr", "local", "bin")
	if err := os.MkdirAll(containerBinPath, 0755); err != nil {
		return nil, fmt.Errorf("create bin directory: %w", err)
	}

	// Write NGINX config files to the context directory.
	if staticDirMap != nil {
		nginxPath := filepath.Join(wrappedContextDirectory, "etc", "nginx", "sites-enabled")
		nginxConfPath := filepath.Join(nginxPath, "default")
		nginxConf := `server {
	listen 8080;
	location / {
		proxy_pass              http://127.0.0.1:8000;
		proxy_set_header        Host $host;
	}
	location ` + staticDirMap.StaticURLPath + ` {
		autoindex on;
		alias ` + staticDirMap.StaticHostDir + ` ;
	}
}`

		if err := os.MkdirAll(nginxPath, 0755); err != nil {
			return nil, fmt.Errorf("create nginx directory: %w", err)
		}

		if err := os.WriteFile(nginxConfPath, []byte(nginxConf), 0644); err != nil {
			return nil, fmt.Errorf("write nginx conf: %w", err)
		}

	}

	venvPath := venvContext.GetPath()

	binariesInVenv, err := os.ReadDir(filepath.Join(venvPath, "bin"))
	if err != nil {
		slog.Warn("no bin directory", slog.String("error", err.Error()))
	} else {
		for _, bin := range binariesInVenv {
			if bin.IsDir() {
				continue
			}

			// files that should not be copied
			if strings.HasPrefix(bin.Name(), "python") ||
				strings.HasPrefix(bin.Name(), "pip") ||
				strings.HasPrefix(bin.Name(), "easy_install") ||
				strings.HasPrefix(bin.Name(), "activate") {
				continue
			}

			slog.Info("replace file", slog.String("file", bin.Name()))

			// replace the shebang if this is a file
			binContent, err := os.ReadFile(filepath.Join(venvPath, "bin", bin.Name()))
			if err != nil {
				slog.Error("cannot read file", slog.String("error", err.Error()))
				continue
			}
			if bytes.HasPrefix(binContent, []byte("#!")) {
				// find the newline
				_, otherText, found := bytes.Cut(binContent, []byte("\n"))
				if found {
					binContent = bytes.Join([][]byte{[]byte("#!/usr/bin/env python3"), otherText}, []byte("\n"))
				}
			}

			err = os.WriteFile(filepath.Join(containerBinPath, bin.Name()), binContent, 0755)
			if err != nil {
				slog.Error("cannot write file", slog.String("error", err.Error()))
			}
		}
	}

	// Copy the context directory to the wrapped context directory.
	if err := cp.Copy(a.Context.Value(sc.ExpandString), path.Join(wrappedContextDirectory, "app")); err != nil {
		return nil, fmt.Errorf("copy context to wrapped context directory: %w", err)
	}

	// Get the runtime dependencies.
	runtimeDependencies := a.RuntimeDependencies.Value(sc.ExpandString)
	var runtimeDependenciesDockerCmd string
	if len(runtimeDependencies) > 0 {
		runtimeDependenciesDockerCmd += "RUN apt update && "
		runtimeDependenciesDockerCmd += "apt install -y " + strings.Join(runtimeDependencies, " ") + " "
		runtimeDependenciesDockerCmd += "&& rm -rf /var/lib/apt/lists/*"
	}

	// Write the Dockerfile.
	dockerfileBuilder := strings.Builder{}
	dockerfileBuilder.WriteString("FROM python:" + pythonVersion + "-slim\n")
	if len(runtimeDependenciesDockerCmd) > 0 {
		dockerfileBuilder.WriteString(runtimeDependenciesDockerCmd + "\n")
	}
	dockerfileBuilder.WriteString("COPY . /\n")
	dockerfileBuilder.WriteString("WORKDIR /app\n")
	dockerfileBuilder.WriteString(dockerCmd + "\n")

	// Call containerized action to build the Docker image.
	step, err := zbaction.ResolveProcedure("zbpack/containerized", zbaction.ProcStepArgs{
		"context":    wrappedContextDirectory,
		"dockerfile": dockerfileBuilder.String(),
		"cache":      strconv.FormatBool(a.Cache.Value(sc.ExpandString)),
	})
	if err != nil {
		return nil, fmt.Errorf("resolve containerized: %w", err)
	}

	return step.Run(ctx, sc)
}
