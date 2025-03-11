package gleam

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Gleam identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeGleam
}

// Match returns true if gleam.toml is found in the source
func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "gleam.toml")
}

func (i *identify) PlanMeta(_ plan.NewPlannerOptions) types.PlanMeta {
	meta := types.PlanMeta{}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
