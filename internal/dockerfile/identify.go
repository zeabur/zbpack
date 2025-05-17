package dockerfile

import (
	"errors"
	"log"
	"os"

	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Dockerfile identifier.
func NewIdentifier() plan.IdentifierV2 {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDocker
}

func (i *identify) Match(ctx plan.MatchContext) bool {
	_, err := FindDockerfile(&FindContext{
		Source:        ctx.Source,
		Config:        ctx.Config,
		SubmoduleName: ctx.SubmoduleName,
	})
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Println("dockerfile: find dockerfile:", err)
		}

		return false
	}

	return true
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(options)
}
