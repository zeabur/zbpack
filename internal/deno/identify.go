package deno

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Deno identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDeno
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "deno.json", "deno.lock", "fresh.gen.ts")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	framework := DetermineFramework(options.Source)
	entry := DetermineEntry(options.Source)
	startCmd := GetStartCommand(options.Source)

	return types.PlanMeta{
		"framework":    string(framework),
		"entry":        entry,
		"startCommand": startCmd,
	}
}

var _ plan.Identifier = (*identify)(nil)
