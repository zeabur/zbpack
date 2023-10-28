package zeaburpack

import (
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// PlanOptions is the options for Plan function.
type PlanOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if a directory is considered as a Go project,
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

// Plan returns the build plan and metadata.
func Plan(opt PlanOptions) (types.PlanType, types.PlanMeta) {
	var src afero.Fs
	if strings.HasPrefix(*opt.Path, "https://github.com") {
		var err error
		src, err = getGitHubSourceFromURL(*opt.Path, *opt.AccessToken)
		if err != nil {
			panic(err)
		}
	} else {
		src = afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	}

	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	UpdateOptionsOnConfig(&opt, config)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             src,
			Config:             config,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
			SubmoduleName:      submoduleName,
		},
		SupportedIdentifiers()...,
	)

	t, m := planner.Plan()
	return t, m
}
