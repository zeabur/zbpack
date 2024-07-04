package rust

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Rust identifier.
func NewIdentifier() plan.ExplainableIdentifier {
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

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		{
			Key:         "BinName",
			Name:        "Rust binary project name",
			Description: "The binary name of the project to deploy",
		},
		{
			Key:         "NeedOpenssl",
			Name:        "Install OpenSSL library",
			Description: "Whether to install the OpenSSL library",
		},
	}

	if _, ok := meta["serverless"]; ok {
		fieldInfo = append(fieldInfo, types.NewServerlessFieldInfo("serverless"))
	}

	return fieldInfo
}

var _ plan.ExplainableIdentifier = (*identify)(nil)
