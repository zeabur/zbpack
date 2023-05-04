package zeaburpack

import (
	"github.com/zeabur/zbpack/internal/plan"
	"github.com/zeabur/zbpack/internal/source"
	"github.com/zeabur/zbpack/pkg/types"
	"strings"
)

type PlanOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// Path is the path to the project directory.
	Path *string

	// Access token for GitHub, only used when Path is a GitHub URL.
	AccessToken *string

	// CustomBuildCommand is a custom build command that will be used instead of the default one.
	CustomBuildCommand *string

	// CustomStartCommand is a custom start command that will be used instead of the default one.
	CustomStartCommand *string

	// OutputDir is the directory where the build artifacts will be stored.
	// Once provided, the service will deploy as static files with nginx.
	OutputDir *string
}

func Plan(opt PlanOptions) (types.PlanType, types.PlanMeta) {

	var src *source.Source
	if strings.HasPrefix(*opt.Path, "https://github.com") {
		var err error
		src, err = getGitHubSourceFromUrl(*opt.Path, *opt.AccessToken)
		if err != nil {
			panic(err)
		}
	} else {
		lSrc := source.NewLocalSource(*opt.Path)
		src = &lSrc
	}

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             src,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
			SubmoduleName:      *opt.SubmoduleName,
		},
	)

	t, m := planner.Plan()
	return t, m
}
