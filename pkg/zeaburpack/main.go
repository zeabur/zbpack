package zeaburpack

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/zeabur/zbpack/internal/nodejs/nextjs"
	"github.com/zeabur/zbpack/internal/static"

	"github.com/spf13/afero"

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

	// HandleBuildFailed is a callback function that will be called when
	// the build failed.
	HandleBuildFailed *func(error)

	// HandleLog is a function that will be called when a log is emitted.
	HandleLog *func(string)

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

	// ProxyRegistry is the registry to be used for the image.
	// See referenceConstructor for more details.
	ProxyRegistry *string
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

	var handleLog func(log string)
	if opt.HandleLog == nil {
		handleLog = func(log string) {
			println(log)
		}
	} else {
		handleLog = func(log string) {
			(*opt.HandleLog)(log)
			println(log)
		}
	}

	var handleBuildFailed func(error)
	if opt.HandleBuildFailed == nil {
		handleBuildFailed = func(err error) {
			println("Build failed: " + err.Error())
		}
	} else {
		handleBuildFailed = func(err error) {
			println("Build failed: " + err.Error())
			(*opt.HandleBuildFailed)(err)
		}
	}

	if strings.HasPrefix(*opt.Path, "https://") {
		println("Build from git repository is not supported yet")
		handleBuildFailed(nil)
		return fmt.Errorf("build from git repository is not supported yet")
	}

	src := afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             src,
			Config:             plan.NewProjectConfigurationFromFs(src),
			SubmoduleName:      *opt.SubmoduleName,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
		},
		SupportedIdentifiers()...,
	)

	t, m := planner.Plan()

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			HandleLog: handleLog,
			planType:  t,
			planMeta:  m,
		},
	)
	if err != nil {
		println("Failed to generate Dockerfile: " + err.Error())
		handleBuildFailed(err)
		return err
	}

	// If the build is interactive, we will print the Dockerfile to the console.
	buildImageHandleLog := &handleLog
	if opt.Interactive != nil && *opt.Interactive {
		buildImageHandleLog = nil
	}

	err = buildImage(
		&buildImageOptions{
			Dockerfile:          dockerfile,
			AbsPath:             *opt.Path,
			UserVars:            *opt.UserVars,
			ResultImage:         *opt.ResultImage,
			HandleLog:           buildImageHandleLog,
			PlainDockerProgress: opt.Interactive == nil || !*opt.Interactive,
			CacheFrom:           opt.CacheFrom,
			ProxyRegistry:       opt.ProxyRegistry,
		},
	)
	if err != nil {
		println("Failed to build image: " + err.Error())
		handleBuildFailed(err)
		return err
	}

	_ = os.RemoveAll(".zeabur")

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkNextJs) && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err = nextjs.TransformServerless(*opt.ResultImage, *opt.Path)
		if err != nil {
			log.Println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["outputDir"] != "" {
		println("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.ResultImage, *opt.Path, m)
		if err != nil {
			println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if opt.Interactive != nil && *opt.Interactive {
		handleLog("\n\033[32mBuild successful\033[0m\n")
		handleLog("\033[90m" + "To run the image, use the following command:" + "\033[0m")
		if t == types.PlanTypeNodejs && m["outputDir"] != "" {
			handleLog("npx serve .zeabur/output/static")
		} else {
			handleLog("docker run -p 8080:8080 -it " + *opt.ResultImage)
		}
	}

	return nil
}
