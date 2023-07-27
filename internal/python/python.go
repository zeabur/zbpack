package python

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Python projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	installCmd := meta["install"]
	startCmd := meta["start"]
	aptDeps := meta["apt-deps"]

	dockerfile := "FROM docker.io/library/python:" + meta["pythonVersion"] + "-slim-buster\n"

	dockerfile += `WORKDIR /app
RUN apt-get update
RUN apt-get install -y ` + aptDeps + `
RUN rm -rf /var/lib/apt/lists/*
COPY . .
RUN ` + installCmd + `
EXPOSE 8080
CMD ` + startCmd

	return dockerfile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Python packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
