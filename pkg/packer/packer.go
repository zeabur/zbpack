// Package packer is the interface definition of packers.
package packer

import (
	"github.com/salamer/zbpack/pkg/plan"
	"github.com/salamer/zbpack/pkg/types"
)

// Packer can identify the plan type and generate a Dockerfile.
type Packer interface {
	plan.Identifier
	GenerateDockerfile(types.PlanMeta) (string, error)
}
