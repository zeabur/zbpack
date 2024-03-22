package plan

import (
	"github.com/spf13/afero"
	zbaction "github.com/zeabur/action"

	"github.com/zeabur/zbpack/pkg/types"
)

// Identifier identifies the plan type and how to get the plan meta.
type Identifier interface {
	PlanType() types.PlanType
	Match(afero.Fs) bool
	PlanMeta(NewPlannerOptions) types.PlanMeta
}

// ActionIdentifier is the interface for identifiers that returns an action.
type ActionIdentifier interface {
	// PlanType returns the type of this identifier, for example, Python.
	PlanType() types.PlanType

	// Planable returns true if the codebase can be planned by this identifier.
	Planable(ctx ProjectContext) bool

	// PlanAction returns the action to execute.
	//
	// If the error is ErrSkipToNext, the executor will skip to the next planner
	// (act like the Match() function).
	PlanAction(ctx ProjectContext) (zbaction.Action, error)
}
