// Package static is the planner of static files
package static

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for static files.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	if meta["framework"] == "hugo" {
		return `FROM hugomods/hugo:exts as builder
WORKDIR /src
COPY . .
RUN hugo --minify

FROM scratch as output
COPY --from=builder /src/public /
`, nil
	}

	if meta["framework"] == "zola" {
		return `FROM ghcr.io/getzola/zola:v` + meta["version"] + ` as builder
WORKDIR /app
COPY . .
RUN ["zola", "build"]

FROM scratch as output
COPY --from=builder /app/public /
`, nil
	}

	if meta["framework"] == "mkdocs" {
		return `FROM squidfunk/mkdocs-material as builder
WORKDIR /docs
COPY . .
RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
RUN mkdocs build

FROM scratch as output
COPY --from=builder /docs/site /
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
