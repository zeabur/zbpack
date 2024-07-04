// Package plan contains the interface about the build plan and the core plan operations.
package plan

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/pkg/types"
)

// Identifier identifies the plan type, how to get the plan meta, and the explanation of the plan meta.
type Identifier interface {
	PlanType() types.PlanType
	Match(afero.Fs) bool
	PlanMeta(NewPlannerOptions) types.PlanMeta
	Explain(meta types.PlanMeta) []types.FieldInfo
}
