package dotnet

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Dotnet identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDotnet
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "Program.cs", "Startup.cs")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	sdkVer, err := DetermineSDKVersion(options.SubmoduleName, options.Source)
	if err != nil {
		panic(err)
	}

	return types.PlanMeta{
		"sdk":        sdkVer,
		"entryPoint": options.SubmoduleName,
	}
}

var _ plan.Identifier = (*identify)(nil)
