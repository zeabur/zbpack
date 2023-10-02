// Package plan is the interface for planners.
package plan

import (
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/zeaburpack"

	"github.com/zeabur/zbpack/pkg/types"
)

// Planner is the interface for planners.
type Planner interface {
	Plan() (types.PlanType, types.PlanMeta)
}

type planner struct {
	NewPlannerOptions

	identifiers []Identifier
}

// NewPlannerOptions is the options for NewPlanner.
type NewPlannerOptions struct {
	Source             afero.Fs
	Config             zeaburpack.ImmutableProjectConfiguration
	SubmoduleName      string
	CustomBuildCommand *string
	CustomStartCommand *string
	OutputDir          *string
}

// NewPlanner creates a new Planner.
func NewPlanner(opt *NewPlannerOptions, identifiers ...Identifier) Planner {
	return &planner{
		NewPlannerOptions: *opt,
		identifiers:       identifiers,
	}
}

func (b planner) Plan() (types.PlanType, types.PlanMeta) {
	for _, identifier := range b.identifiers {
		if identifier.Match(b.Source) {
			return identifier.PlanType(), identifier.PlanMeta(b.NewPlannerOptions)
		}
	}

	return types.PlanTypeStatic, types.PlanMeta{}
}
