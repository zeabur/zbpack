// Package plan contains the interface about the build plan and the core plan operations.
package plan

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/pkg/types"
)

// Identifier identifies the plan type and how to get the plan meta.
type Identifier interface {
	PlanType() types.PlanType
	Match(afero.Fs) bool
	PlanMeta(NewPlannerOptions) types.PlanMeta
}

// ExplainableIdentifier is an identifier that can explain the fields in the plan.
type ExplainableIdentifier interface {
	Identifier
	Explain(meta types.PlanMeta) []types.FieldInfo
}
