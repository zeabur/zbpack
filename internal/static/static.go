// Package static is the planner of static files
package static

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for static files.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	var dockerfile string

	switch meta["framework"] {
	case "hugo":
		dockerfile = `FROM hugomods/hugo:exts AS builder
WORKDIR /src
COPY . .
RUN hugo --minify

FROM scratch AS output
COPY --from=builder /src/public /
`
	case "zola":
		dockerfile = `FROM ghcr.io/getzola/zola:v` + meta["version"] + ` AS builder
WORKDIR /app
COPY . .
RUN ["zola", "build"]

FROM scratch AS output
COPY --from=builder /app/public /
`

	case "mkdocs":
		dockerfile = `FROM squidfunk/mkdocs-material AS builder
WORKDIR /docs
COPY . .
RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
RUN mkdocs build

FROM scratch AS output
COPY --from=builder /docs/site /
`

	default:
		dockerfile = `FROM scratch AS output
COPY . /
`
	}

	// We run it with caddy for Containerized mode.
	if serverless, ok := meta["serverless"]; ok && serverless != "true" {
		caddy := `FROM zeabur/caddy-static AS runtime
COPY --from=output / /usr/share/caddy
`

		dockerfile += "\n" + caddy
	}

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
