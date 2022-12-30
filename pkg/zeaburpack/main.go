package zeaburpack

import (
	"os"
	"path"
	"strings"

	. "github.com/zeabur/zbpack/pkg/types"
)

type BuildOptions struct {

	// SubmoduleName is the of the submodule to build.
	// For example, if directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// HandlePlanDetermined is a callback function that will be called when
	// the build plan is determined.
	HandlePlanDetermined *func(PlanType, PlanMeta)

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

	var handlePlanDetermined func(planType PlanType, planMeta PlanMeta)
	if opt.HandlePlanDetermined == nil {
		handlePlanDetermined = func(planType PlanType, planMeta PlanMeta) {
			PrintPlanAndMeta(planType, planMeta, handleLog)
		}
	} else {
		handlePlanDetermined = func(planType PlanType, planMeta PlanMeta) {
			PrintPlanAndMeta(planType, planMeta, handleLog)
			(*opt.HandlePlanDetermined)(planType, planMeta)
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

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			AbsPath:              *opt.Path,
			SubmoduleName:        *opt.SubmoduleName,
			HandleLog:            handleLog,
			HandlePlanDetermined: handlePlanDetermined,
		},
	)
	if err != nil {
		handleBuildFailed(err)
		return err
	}

	err = buildImage(
		&BuildImageOptions{
			Dockerfile:  dockerfile,
			AbsPath:     *opt.Path,
			UserVars:    *opt.UserVars,
			ResultImage: *opt.ResultImage,
			HandleLog:   handleLog,
		},
	)
	if err != nil {
		handleBuildFailed(err)
		return err
	}

	return nil
}
