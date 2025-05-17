// Package packer is the interface definition of packers.
package packer

import (
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// Packer can identify the plan type and generate a Dockerfile.
type Packer interface {
	plan.Identifier
	GenerateDockerfile(types.PlanMeta) (string, error)
}

// V2 can identify the plan type and generate a Dockerfile.
type V2 interface {
	plan.IdentifierV2
	GenerateDockerfile(types.PlanMeta) (string, error)
}

// WrapV2 wraps a Packer to a V2 packer.
func WrapV2(p Packer) V2 {
	return &packerV2Wrapper{
		Packer: p,
	}
}

type packerV2Wrapper struct {
	Packer
}

func (p *packerV2Wrapper) Match(ctx plan.MatchContext) bool {
	return p.Packer.Match(ctx.Source)
}
