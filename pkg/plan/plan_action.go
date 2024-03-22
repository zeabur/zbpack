package plan

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/afero"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/pkg/types"
)

// ErrSkipToNext indicates the executor to skip to the next planner.
var ErrSkipToNext = errors.New("plan: skip to next")

// AggregatedActionPlanner is the aggregated planner that finds the best matched
// action planner
type AggregatedActionPlanner struct {
	identifiers []ActionIdentifier
}

// ProjectContext is the project context for planning.
type ProjectContext struct {
	Source afero.Fs
	Config ImmutableProjectConfiguration

	// SubmoduleName is the name of the submodule.
	// In CLI, it is the name of the source directory.
	// In Zeabur, it is the name of the project.
	SubmoduleName string
}

// NewAggregatedActionPlanner creates a new AggregatedActionPlanner.
//
//		NewAggregatedActionPlanner(
//		    PythonIdentifier,
//		    NodeIdentifier,
//		    GoIdentifier,
//	        StaticIdentifier, // Act like fallback! It should accept all kind of projects.
//		)
//
// The first identifier has the highest priority.
func NewAggregatedActionPlanner(identifiers ...ActionIdentifier) *AggregatedActionPlanner {
	return &AggregatedActionPlanner{
		identifiers: identifiers,
	}
}

// PlanAction plans the action to execute with the given identifiers, according to the context.
func (ap *AggregatedActionPlanner) PlanAction(context ProjectContext) (types.PlanType, zbaction.Action, error) {
	for _, identifier := range ap.identifiers[:len(ap.identifiers)-1] {
		if !identifier.Planable(context) {
			continue
		}

		plannedAction, err := identifier.PlanAction(context)
		if err != nil {
			if !errors.Is(err, ErrSkipToNext) {
				log.Printf("failed to plan this project with %s: %v\n", identifier, err)
			}

			continue
		}

		return identifier.PlanType(), plannedAction, nil
	}

	// Run the last (fallback) identifier – it always runs no matter Match() returns true or false.
	finalIdentifier := ap.identifiers[len(ap.identifiers)-1]
	plannedAction, err := finalIdentifier.PlanAction(context)
	if err != nil {
		if errors.Is(err, ErrSkipToNext) {
			return "", plannedAction, fmt.Errorf("developer's matter: no action handles this case")
		}
		return finalIdentifier.PlanType(), plannedAction, err
	}

	return finalIdentifier.PlanType(), plannedAction, nil
}
