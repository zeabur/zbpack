package php

import (
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new PHP identifier.
func NewIdentifier() plan.ExplainableIdentifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypePHP
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "index.php", "composer.json")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	config := options.Config

	server := plan.Cast(config.Get(ConfigLaravelOctaneServer), castOctaneServer).TakeOr("")

	framework := DetermineProjectFramework(options.Source)
	phpVersion := GetPHPVersion(config, options.Source)
	deps := DetermineAptDependencies(options.Source, server)
	exts := DeterminePHPExtensions(options.Source)
	app, property := DetermineApplication(options.Source)
	buildCommand := DetermineBuildCommand(options.Source, options.Config, options.CustomBuildCommand)
	startCommand := DetermineStartCommand(options.Config, options.CustomStartCommand)

	// Some meta will be added to the plan dynamically later.
	meta := types.PlanMeta{
		"framework":    string(framework),
		"phpVersion":   phpVersion,
		"deps":         strings.Join(deps, " "),
		"exts":         strings.Join(exts, " "),
		"app":          string(app),
		"property":     PropertyToString(property),
		"buildCommand": buildCommand,
		"startCommand": startCommand,
	}

	if framework == types.PHPFrameworkLaravel && server != "" {
		meta["octaneServer"] = server
	}

	return meta
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		types.NewFrameworkFieldInfo("framework", types.PlanTypePHP, meta["framework"]),
		{
			Key:         "phpVersion",
			Name:        "PHP Version",
			Description: "The version of PHP for building in the source code",
		},
		{
			Key:         "deps",
			Name:        "Dependencies",
			Description: "The runtime dependencies required by the project.",
		},
		{
			Key:         "exts",
			Name:        "PHP Extensions",
			Description: "The PHP extensions required by the project.",
		},
		{
			Key:         "app",
			Name:        "Application",
			Description: "The type of application. It decides the Nginx configuration to use.",
		},
		types.NewBuildCmdFieldInfo("buildCommand"),
		types.NewStartCmdFieldInfo("startCommand"),
	}

	if _, ok := meta["octaneServer"]; ok {
		fieldInfo = append(fieldInfo, types.FieldInfo{
			Key:         "octaneServer",
			Name:        "Laravel Octane server type",
			Description: "The server type used by Laravel Octane such as swoole and roadrunner.",
		})
	}

	// wip: property (verbose)

	return fieldInfo
}

var _ plan.ExplainableIdentifier = (*identify)(nil)
