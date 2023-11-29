// Package static is the planner of static files
package static

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for static files.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {

	if meta["framework"] == "hugo" {
		return `FROM klakegg/hugo:ubuntu as builder
WORKDIR /src
RUN apt-get update && apt-get install -y git
COPY . .
RUN hugo --minify

FROM scratch as output
COPY --from=builder /src/public /
`, nil
	}

	dockerfile := `FROM scratch as output
COPY . /
`

	return dockerfile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new static packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
