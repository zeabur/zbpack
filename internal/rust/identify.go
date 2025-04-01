package rust

import (
	"github.com/spf13/afero"

	"github.com/salamer/zbpack/internal/utils"
	"github.com/salamer/zbpack/pkg/plan"
	"github.com/salamer/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Rust identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeRust
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "Cargo.toml")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:           options.Source,
			SubmoduleName: options.SubmoduleName,
			Config:        options.Config,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
