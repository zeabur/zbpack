package python

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Python projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	installCmd := meta["install"]
	buildCmd := meta["build"]
	startCmd := meta["start"]
	aptDeps := meta["apt-deps"]
	staticMeta := staticInfoFromMeta(meta)

	dockerfile := "FROM docker.io/library/python:" + meta["pythonVersion"] + "-slim\n"

	if staticMeta.NginxEnabled() {
		dockerfile += `WORKDIR /app
RUN apt-get update && apt-get install -y ` + aptDeps + ` \
&& rm /etc/nginx/sites-enabled/default \
&& echo "\
server { \
        listen 8080; \
        location / { \
			proxy_pass              http://127.0.0.1:8000; \
			proxy_set_header        Host \$host; \
		} \
		location ` + staticMeta.StaticURLPath + `{ \
			autoindex on; \
			alias ` + staticMeta.StaticHostDir + ` ; \` + `
		} \
    }"> /etc/nginx/conf.d/default.conf ` + ` && rm -rf /var/lib/apt/lists/*
` + installCmd + `
COPY . .
EXPOSE 8080
CMD ` + startCmd
	} else {
		dockerfile += `
WORKDIR /app
RUN apt-get update
RUN apt-get install -y ` + aptDeps + `
RUN rm -rf /var/lib/apt/lists/*
` + installCmd + `
COPY . .
` + buildCmd + `
EXPOSE 8080
CMD ` + startCmd
	}

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
