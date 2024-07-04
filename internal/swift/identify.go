package swift

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Python identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeSwift
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "Package.swift")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(GetMetaOptions{Src: options.Source, Config: options.Config})
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	if _, ok := meta["framework"]; ok {
		return []types.FieldInfo{
			types.NewFrameworkFieldInfo("framework", types.PlanTypeSwift, meta["framework"]),
		}
	}

	return []types.FieldInfo{}
}

var _ plan.ExplainableIdentifier = (*identify)(nil)
