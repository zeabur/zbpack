// Package golang is the planner for Golang projects.
package golang

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Golang projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	buildStage := `FROM docker.io/library/golang:` + meta["goVersion"] + `-alpine as builder
RUN mkdir /src
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . /src/
ENV CGO_ENABLED=0
RUN go build -o ./bin/server ` + meta["entry"]

	runtimeStage := `FROM alpine as runtime
COPY --from=builder /src/bin/server /bin/server
CMD ["/bin/server"]`

	exportStage := `FROM scratch
COPY --from=builder /src/bin/server main
`

	if meta["serverless"] == "true" {
		return buildStage + "\n" + exportStage, nil
	}

	return buildStage + "\n" + runtimeStage, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Golang packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
