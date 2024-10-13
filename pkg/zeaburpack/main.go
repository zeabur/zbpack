package zeaburpack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/dockerfiles"
	"github.com/zeabur/zbpack/pkg/plan"
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

	// HandleLog is a function that will be called when a log is emitted.
	HandleLog *func(string)

	// LogWriter is the writer to write the buildkit logs to.
	// If not provided, the logs will be written to os.Stderr.
	LogWriter io.Writer

	// Path is the path to the project directory.
	Path *string

	// ResultImage is the name of the image that will be built.
	ResultImage *string

	// UserVars is a map of user variables that will be used in the Dockerfile.
	UserVars *map[string]string

	// Interactive is a flag to indicate if the build should be interactive.
	Interactive *bool

	// CustomBuildCommand is a custom build command that will be used instead of the default one.
	CustomBuildCommand *string

	// CustomStartCommand is a custom start command that will be used instead of the default one.
	CustomStartCommand *string

	// OutputDir is the directory where the build artifacts will be stored.
	// Once provided, the service will deploy as static files with nginx.
	OutputDir *string

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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if opt.Path == nil || *opt.Path == "" {
		opt.Path = &wd
	} else if !strings.HasPrefix(*opt.Path, "/") {
		p := path.Join(wd, *opt.Path)
		opt.Path = &p
	}

	if opt.SubmoduleName == nil {
		opt.SubmoduleName = lo.ToPtr("")
	}

	if opt.ResultImage == nil || *opt.ResultImage == "" {
		opt.ResultImage = lo.ToPtr(path.Base(*opt.Path))
	}

	*opt.ResultImage = strings.ToLower(*opt.ResultImage)
	*opt.ResultImage = strings.ReplaceAll(*opt.ResultImage, "_", "-")

	if opt.UserVars == nil {
		emptyUserVars := make(map[string]string)
		opt.UserVars = &emptyUserVars
	}

	if os.Getenv("REGISTRY") != "" {
		opt.ProxyRegistry = lo.ToPtr(os.Getenv("REGISTRY"))
	}

	var handleLog func(log string)
	if opt.HandleLog == nil {
		handleLog = func(log string) {
			println(log)
		}
	} else {
		handleLog = *opt.HandleLog
	}

	if strings.HasPrefix(*opt.Path, "https://") {
		println("Build from git repository is not supported yet")
		return fmt.Errorf("build from git repository is not supported yet")
	}

	src := afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	UpdateOptionsOnConfig(opt, config)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             src,
			Config:             config,
			SubmoduleName:      submoduleName,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
		},
		SupportedIdentifiers(config)...,
	)

	t, m := planner.Plan()

	PrintPlanAndMeta(t, m, handleLog)

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	if m["zeaburImage"] == "" {
		return fmt.Errorf("zeaburImage is not set")
	}

	dockerfileContent, err := dockerfiles.GetDockerfileContent(m["zeaburImage"])
	if err != nil {
		return fmt.Errorf("get Dockerfile content: %w", err)
	}
	injectedDockerfileContent := InjectDockerfile(string(dockerfileContent), opt.ProxyRegistry, *opt.UserVars)

	builder := ImageBuilder{
		Path:              *opt.Path,
		PlanMeta:          m,
		ResultImage:       *opt.ResultImage,
		DockerfileContent: injectedDockerfileContent,
		Stage:             m["zeaburImageStage"],
		BuildArgs:         m,
		LogWriter:         opt.LogWriter,
	}
	artifact, err := builder.BuildImage(context.Background())
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	if opt.Interactive != nil && *opt.Interactive {
		handleLog("\n\033[32mBuild successful\033[0m\n")

		if tar, ok := artifact.GetDockerTar(); ok {
			handleLog("\033[90m" + "The Docker image has been saved in " + tar + "\033[0m")
		}

		if dotZeaburDirectory, ok := artifact.GetDotZeaburDirectory(); ok {
			handleLog("\033[90m" + "The compiled serverless function has been saved in " + dotZeaburDirectory + "\033[0m")
		}
	}

	return nil
}
