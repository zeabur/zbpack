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
func NewIdentifier() plan.Identifier {
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
	phpVersion := GetPHPVersion(options.Source)
	deps := DetermineAptDependencies(options.Source, server)
	exts := DeterminePHPExtensions(options.Source)
	app, property := DetermineApplication(options.Source)
	startCommand := DetermineStartCommand(options.Config, options.CustomStartCommand)

	// Some meta will be added to the plan dynamically later.
	meta := types.PlanMeta{
		"framework":    string(framework),
		"phpVersion":   phpVersion,
		"deps":         strings.Join(deps, " "),
		"exts":         strings.Join(exts, " "),
		"app":          string(app),
		"property":     PropertyToString(property),
		"startCommand": startCommand,
	}

	if framework == types.PHPFrameworkLaravel && server != "" {
		meta["octaneServer"] = server
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
