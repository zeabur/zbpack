package golang

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Golang identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeGo
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "go.mod")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:           options.Source,
			Config:        options.Config,
			SubmoduleName: options.SubmoduleName,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
