package gleam

import (
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

type gleamPlanContext struct {
	Src        afero.Fs
	Config     plan.ImmutableProjectConfiguration
	Serverless optional.Option[bool]
}

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

func getServerless(ctx *gleamPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
}

func (i *identify) PlanMeta(opt plan.NewPlannerOptions) types.PlanMeta {
	meta := types.PlanMeta{}

	ctx := &gleamPlanContext{
		Src:    opt.Source,
		Config: opt.Config,
	}

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = "true"
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
