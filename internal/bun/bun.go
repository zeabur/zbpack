// Package bun generates the Dockerfile for Bun projects.
// It is currently the wrapper of Node.js planner.
package bun

import (
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Bun projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return nodejs.GenerateDockerfile(meta)
}

type pack struct {
	*identify
}

// NewPacker returns a new Node.js packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
