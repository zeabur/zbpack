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

// IdentifierV2 is the new version of Identifier.
type IdentifierV2 interface {
	PlanType() types.PlanType
	Match(MatchContext) bool
	PlanMeta(NewPlannerOptions) types.PlanMeta
}

// MatchContext is the context for matching.
type MatchContext struct {
	Source        afero.Fs
	Config        ImmutableProjectConfiguration
	SubmoduleName string
}

// WrapV2 wraps an Identifier to an IdentifierV2.
func WrapV2(i Identifier) IdentifierV2 {
	return &identifierV2Wrapper{
		Identifier: i,
	}
}

type identifierV2Wrapper struct {
	Identifier
}

func (i *identifierV2Wrapper) Match(ctx MatchContext) bool {
	return i.Identifier.Match(ctx.Source)
}
