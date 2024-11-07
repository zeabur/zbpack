package zeaburpack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/codeclysm/extract/v3"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/transformer"
	"github.com/zeabur/zbpack/pkg/types"
)

// BuildOptions is the options for the Build function.
type BuildOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// HandlePlanDetermined is a callback function that will be called when
	// the build plan is determined.
	HandlePlanDetermined *func(types.PlanType, types.PlanMeta)

	// LogWriter is a [io.Writer] that will be written when a log is emitted.
	// nil to use the default log writer.
	LogWriter io.Writer

	// Path is the path to the project directory.
	Path *string

	// ResultImage is the name of the image that will be built.
	ResultImage *string

	// UserVars is a map of user variables that will be used in the Dockerfile.
	UserVars *map[string]string

	// Interactive is a flag to indicate if the build should be interactive.
	Interactive *bool

	CacheFrom *string
	CacheTo   *string

	// ProxyRegistry is the registry to be used for the image.
	// See referenceConstructor for more details.
	ProxyRegistry *string

	// PushImage is a flag to indicate if the image should be pushed to the registry.
	PushImage bool
}

// Build will analyze the project, determine the plan and build the image.
func Build(opt *BuildOptions) error {
	// clean up the buildkit output directory after the build
	defer func() {
		_ = os.RemoveAll(path.Join(os.TempDir(), "zbpack/buildkit"))
	}()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if opt.LogWriter == nil {
		opt.LogWriter = os.Stderr
	}

	if opt.Path == nil || *opt.Path == "" {
		opt.Path = &wd
	} else if !strings.HasPrefix(*opt.Path, "/") {
		p := path.Join(wd, *opt.Path)
		opt.Path = &p
	}

	if opt.SubmoduleName == nil {
		emptySubmoduleName := ""
		opt.SubmoduleName = &emptySubmoduleName
	}

	if opt.ResultImage == nil || *opt.ResultImage == "" {
		img := path.Base(*opt.Path)
		opt.ResultImage = &img
	}

	*opt.ResultImage = strings.ToLower(*opt.ResultImage)
	*opt.ResultImage = strings.ReplaceAll(*opt.ResultImage, "_", "-")

	if opt.UserVars == nil {
		emptyUserVars := make(map[string]string)
		opt.UserVars = &emptyUserVars
	}

	if strings.HasPrefix(*opt.Path, "https://") {
		opt.Log("Build from git repository is not supported yet\n")
		return fmt.Errorf("build from git repository is not supported yet")
	}

	src := afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        src,
			Config:        config,
			SubmoduleName: submoduleName,
		},
		SupportedIdentifiers(config)...,
	)

	t, m := planner.Plan()

	PrintPlanAndMeta(t, m, opt.LogWriter)

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			planType: t,
			planMeta: m,
		},
	)
	if err != nil {
		opt.Log("Failed to generate Dockerfile: %s\n", err)
		return err
	}

	// Remove .zeabur directory if exists
	_ = os.RemoveAll(path.Join(*opt.Path, ".zeabur"))

	// Inject dockerfile to contain the variables, registry, etc.
	newDockerfile := InjectDockerfile(dockerfile, opt.ProxyRegistry, *opt.UserVars)

	err = buildImage(
		&buildImageOptions{
			PlanType: t,
			PlanMeta: m,

			Dockerfile:          newDockerfile,
			AbsPath:             *opt.Path,
			UserVars:            *opt.UserVars,
			PlainDockerProgress: opt.Interactive == nil || !*opt.Interactive,

			ResultImage: *opt.ResultImage,
			PushImage:   opt.PushImage,

			CacheFrom: opt.CacheFrom,
			CacheTo:   opt.CacheTo,

			LogWriter: opt.LogWriter,
		},
	)
	if err != nil {
		opt.Log("Failed to build image: %s\n", err)
		return err
	}

	dockerBuildOutput := path.Join(os.TempDir(), "zbpack/buildkit")
	// decompress TAR to the output directory
	func() {
		if err := os.MkdirAll(dockerBuildOutput, 0o755); err != nil {
			println("Failed to create output directory: " + err.Error())
			return
		}

		// decompress the given TAR file to the output directory
		tarFile, err := os.Open(ServerlessTarPath)
		if err != nil {
			if m["serverless"] == "true" {
				opt.Log("Failed to open TAR file: %s\n", err)
			}
			return
		}
		defer func(tarFile *os.File) {
			_ = tarFile.Close()

			// clean up TAR file
			_ = os.Remove(ServerlessTarPath)
		}(tarFile)

		err = extract.Tar(context.TODO(), tarFile, dockerBuildOutput, func(filename string) string {
			switch filename {
			case ".git", ".github", ".vscode", ".idea", ".gitignore",
				"Dockerfile", "LICENSE", "README.md", "Makefile",
				".pre-commit-config.yaml":
				return "" // skip these files
			default:
				return filename
			}
		})
		if err != nil {
			opt.Log("Failed to decompress TAR: %s\n", err)
			return
		}
	}()

	opt.Log("Transforming build output ...\n")
	err = transformer.Transform(&transformer.Context{
		PlanType:     t,
		PlanMeta:     m,
		BuildkitPath: dockerBuildOutput,
		AppPath:      *opt.Path,
		PushImage:    opt.PushImage,
		ResultImage:  *opt.ResultImage,
		LogWriter:    opt.LogWriter,
	})
	if err != nil {
		opt.Log("Failed to transform build output: %s\n", err)
		return fmt.Errorf("transform build output: %w", err)
	}

	if opt.Interactive != nil && *opt.Interactive {
		opt.Log("\n\033[32mBuild successful\033[0m\n")
		if m["serverless"] == "true" {
			opt.Log("\033[90m" + "The compiled serverless function has been saved in the .zeabur directory." + "\033[0m\n")
		} else {
			opt.Log("\033[90m" + "To run the image, use the following command:" + "\033[0m\n")
			if m["outputDir"] != "" && m["serverless"] == "true" {
				opt.Log("npx serve .zeabur/output/static\n")
			} else {
				opt.Log("docker run -p 8080:8080 -e PORT=8080 -it %s\n", *opt.ResultImage)
			}
		}
	}

	return nil
}

// Log writes a log message to the log writer.
//
// It passes the parameters to [fmt.Fprintf] internally.
func (opt *BuildOptions) Log(msg string, args ...any) {
	_, _ = fmt.Fprintf(opt.LogWriter, msg, args...)
}
