// Package deno is the planner for Deno projects.
package deno

import (
	"github.com/salamer/zbpack/pkg/packer"
	"github.com/salamer/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Deno projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	framework := meta["framework"]
	entry := meta["entry"]
	startCmd := meta["startCommand"]

	dockerfile := `FROM docker.io/denoland/deno
WORKDIR /app
COPY . .
EXPOSE 8080
RUN deno cache ` + entry

	switch framework {
	case string(types.DenoFrameworkFresh):
		dockerfile += `
CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-run", "` + entry + `"]`
	case string(types.DenoFrameworkNone):
		if startCmd == "" {
			dockerfile += `
CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-run", "` + entry + `"]`
		} else {
			dockerfile += `
CMD ["deno", "task", "start"]`
		}
	}
	return dockerfile, nil
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
