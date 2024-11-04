// Package golang is the planner for Golang projects.
package golang

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Golang projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	cgoEnvSegment := "ENV CGO_ENABLED=0\n"
	if meta["cgo"] == "true" {
		cgoEnvSegment = "ENV CGO_ENABLED=1\n"
	}

	dependencySegment := ""
	if meta["cgo"] == "true" {
		dependencySegment = "RUN apk add --no-cache build-base cmake\n"
	}

	buildCommandSegment := ""
	if meta["buildCommand"] != "" {
		buildCommandSegment = `RUN ` + meta["buildCommand"] + "\n"
	}

	buildStage := `FROM docker.io/library/golang:` + meta["goVersion"] + `-alpine AS builder
RUN mkdir /src
WORKDIR /src
` + dependencySegment + `
COPY go.mod go.sum* ./
RUN go mod download
COPY . /src/
` + cgoEnvSegment + buildCommandSegment + `
RUN go build -o ./bin/server ` + meta["entry"]

	runtimeStage := `FROM alpine AS runtime
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
