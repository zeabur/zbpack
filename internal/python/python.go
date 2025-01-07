package python

import (
	"strconv"

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
	serverless := meta["serverless"]
	pyVer := meta["pythonVersion"]

	if meta["framework"] == string(types.PythonFrameworkReflex) {
		return `FROM python:` + pyVer + `
RUN apt-get update -y && apt-get install -y caddy && rm -rf /var/lib/apt/lists/*
WORKDIR /app
RUN cat > Caddyfile <<EOF
:8080
encode gzip
@backend_routes path /_event/* /ping /_upload /_upload/*
handle @backend_routes {
	reverse_proxy localhost:8000
}
root * /srv
route {
	try_files {path} {path}/ /404.html
	file_server
}
EOF

COPY . .
` + installCmd + `
` + buildCmd + `
STOPSIGNAL SIGKILL
CMD ` + startCmd, nil
	}

	if serverless == "true" {
		return `FROM docker.io/library/python:` + pyVer + `-slim AS builder
WORKDIR /app
COPY . .
` + installCmd + `
` + buildCmd + `

FROM scratch AS output
COPY --from=builder /usr/local/lib/python` + pyVer + `/site-packages /.site-packages
COPY --from=builder /app /
`, nil
	}

	dockerfile := "FROM docker.io/library/python:" + pyVer + "-slim\n"
	dockerfile += `WORKDIR /app
RUN apt-get update && apt-get install -y ` + aptDeps + " && rm -rf /var/lib/apt/lists/*\n"

	// if selenium is required, we install chromium
	// https://github.com/SeleniumHQ/docker-selenium/blob/f39a9da86f635b21d6dff0572e7713dc80c20d69/NodeChrome/Dockerfile#L17C1-L32C50
	if meta["selenium"] == "true" {
		dockerfile += `RUN apt update -y \
		&& apt install -y curl \
		&& (curl https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor | tee /etc/apt/trusted.gpg.d/google.gpg >/dev/null) \
		&& (echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list) \
		&& apt update -y \
		&& apt install -y google-chrome-stable \
		&& rm -f /etc/apt/sources.list.d/google-chrome.list \
		&& rm -rf /var/lib/apt/lists/* /var/cache/apt/*` + "\n"
	}

	if staticMeta.NginxEnabled() {
		dockerfile += `RUN rm /etc/nginx/sites-enabled/default \
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
}"> /etc/nginx/conf.d/default.conf` + "\n"
	}

	dockerfile += "COPY . .\n" + installCmd + "\n" + buildCmd + `
EXPOSE 8080
CMD ["/bin/bash", "-c", ` + strconv.Quote(startCmd) + `]`

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
