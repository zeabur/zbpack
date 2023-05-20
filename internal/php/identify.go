package php

import (
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
	framework := DetermineProjectFramework(options.Source)
	phpVersion := GetPHPVersion(options.Source)

	return types.PlanMeta{
		"framework":  string(framework),
		"phpVersion": phpVersion,
	}
}

var _ plan.Identifier = (*identify)(nil)
