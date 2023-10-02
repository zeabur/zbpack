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

	framework := DetermineProjectFramework(options.Source)
	phpVersion := GetPHPVersion(options.Source)
	deps := DetermineAptDependencies(options.Source)
	app, property := DetermineApplication(options.Source)

	// Some meta will be added to the plan dynamically later.
	meta := types.PlanMeta{
		"framework":  string(framework),
		"phpVersion": phpVersion,
		"deps":       strings.Join(deps, " "),
		"app":        string(app),
		"property":   PropertyToString(property),
	}

	if framework == types.PHPFrameworkLaravel {
		if runner := config.GetString("laravel.octane.server"); runner != "" {
			meta["octaneServer"] = runner
		}
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
