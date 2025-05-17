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
