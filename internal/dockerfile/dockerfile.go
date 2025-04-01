package dockerfile

import (
	"github.com/salamer/zbpack/pkg/packer"
	"github.com/salamer/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for static files.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return meta["content"], nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Deno packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
