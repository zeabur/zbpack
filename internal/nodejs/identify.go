package nodejs

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
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
			Src:            options.Source,
			Config:         options.Config,
			CustomBuildCmd: options.CustomBuildCommand,
			CustomStartCmd: options.CustomStartCommand,
			OutputDir:      options.OutputDir,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
