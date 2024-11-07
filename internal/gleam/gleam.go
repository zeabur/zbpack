// Package gleam is the packer for Gleam projects.
package gleam

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Gleam projects.
func GenerateDockerfile(m types.PlanMeta) (string, error) {
	dockerfile := `FROM ghcr.io/gleam-lang/gleam:v1.3.2-erlang-alpine
RUN apk add --no-cache elixir
RUN mix local.hex --force
RUN mix local.rebar --force
COPY . /build/
RUN cd /build \
  && gleam export erlang-shipment \
  && mv build/erlang-shipment /app \
  && rm -r /build`

	if m["serverless"] == "true" {
		dockerfile += `
FROM scratch AS output
COPY --from=0 /app /
`
	} else {
		dockerfile += `
WORKDIR /app
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["run"]
`
	}

	return dockerfile, nil
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
