// Package static is the planner of static files
package static

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for static files.
func GenerateDockerfile(_ types.PlanMeta) (string, error) {
	dockerfile := `FROM docker.io/library/nginx:alpine as runtime
WORKDIR /usr/share/nginx/html/static
COPY . .
RUN echo "server { listen 8080; root /usr/share/nginx/html/static; }"> /etc/nginx/conf.d/default.conf
EXPOSE 8080`

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
