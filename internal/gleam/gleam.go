// Package gleam is the packer for Gleam projects.
package gleam

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Elixir projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return `FROM ghcr.io/gleam-lang/gleam:v1.0.0-erlang-alpine
COPY . /build/
RUN cd /build \
  && gleam export erlang-shipment \
  && mv build/erlang-shipment /app \
  && rm -r /build
WORKDIR /app
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["run"]`, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Dotnet packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
