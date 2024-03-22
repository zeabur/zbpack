package zeaburpack

import (
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/golang"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// PlanActionOptions is the options for PlanAction.
type PlanActionOptions struct {
	Source afero.Fs

	// SubmoduleName is the name of the submodule.
	// If not provided, it is default to the name of the directory.
	SubmoduleName optional.Option[string]
}

// PlanAction returns the planned action.
func PlanAction(opt PlanActionOptions) (types.PlanType, zbaction.Action, error) {
	submoduleName := opt.SubmoduleName.TakeOr(opt.Source.Name())
	config := plan.NewProjectConfigurationFromFs(opt.Source, submoduleName)

	planner := plan.NewAggregatedActionPlanner(
		golang.NewActionIdentifier(),
		python.NewActionIdentifier(),
	)

	planType, plannedAction, err := planner.PlanAction(plan.ProjectContext{
		Source:        opt.Source,
		Config:        config,
		SubmoduleName: submoduleName,
	})

	return planType, plannedAction, err
}
