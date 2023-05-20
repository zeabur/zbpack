package ruby

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Ruby identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeRuby
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "Gemfile")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	rubyVersion := DetermineRubyVersion(options.Source)
	framework := DetermineRubyFramework(options.Source)

	return types.PlanMeta{
		"rubyVersion": rubyVersion,
		"framework":   string(framework),
	}
}

var _ plan.Identifier = (*identify)(nil)
