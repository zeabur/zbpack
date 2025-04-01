package nodejs

import (
	"github.com/spf13/afero"

	"github.com/salamer/zbpack/internal/utils"
	"github.com/salamer/zbpack/pkg/plan"
	"github.com/salamer/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new NodeJS identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeNodejs
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "package.json")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:    options.Source,
			Config: options.Config,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
