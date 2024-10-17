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

	if strings.HasPrefix(*opt.Path, "https://") {
		_, _ = fmt.Fprintln(opt.LogWriter, "Build from git repository is not supported yet")
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

	dockerfileContent, err := GetDockerfileContent(t, m, opt.ProxyRegistry, *opt.UserVars)
	if err != nil {
		return fmt.Errorf("get Dockerfile content: %w", err)
	}

	builder := ImageBuilder{
		Path:              *opt.Path,
		PlanMeta:          m,
		ResultImage:       *opt.ResultImage,
		DockerfileContent: dockerfileContent,
		Stage:             m["zeaburImageStage"],
		BuildArgs:         m,
		LogWriter:         opt.LogWriter,
	}
	artifact, err := builder.BuildImage(context.Background())
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	if opt.Interactive != nil && *opt.Interactive {
		_, _ = fmt.Fprint(opt.LogWriter, "\n\033[32mBuild successful\033[0m\n")

		if tar, ok := artifact.GetDockerTar(); ok {
			_, _ = fmt.Fprint(opt.LogWriter, "\033[90m"+"The Docker image has been saved in "+tar+"\033[0m")
		}

		if dotZeaburDirectory, ok := artifact.GetDotZeaburDirectory(); ok {
			_, _ = fmt.Fprint(opt.LogWriter, "\033[90m"+"The compiled serverless function has been saved in "+dotZeaburDirectory+"\033[0m")
		}
	}

	return nil
}

// GetDockerfileContent returns the content of the Dockerfile
// of the given plan type and plan meta.
func GetDockerfileContent(
	t types.PlanType,
	m types.PlanMeta,
	proxyRegistry *string,
	userVars map[string]string,
) (string, error) {
	var dockerfileContent string

	if t == types.PlanTypeDocker {
		dockerfileContent = m["content"]
	}
	if m["zeaburImage"] != "" {
		dockerfileContentBytes, err := dockerfiles.GetDockerfileContent(m["zeaburImage"])
		if err != nil {
			return "", fmt.Errorf("get Dockerfile content: %w", err)
		}

		dockerfileContent = string(dockerfileContentBytes)
	}

	if dockerfileContent == "" {
		return "", fmt.Errorf("no Dockerfile content found")
	}

	return InjectDockerfile(string(dockerfileContent), proxyRegistry, userVars), nil
}
