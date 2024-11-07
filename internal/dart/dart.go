// Package dart provides a Dart packer.
package dart

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Dart projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	build := meta["build"]

	if meta["framework"] == "flutter" {
		dockerfile := `FROM ubuntu:latest
RUN apt-get update && apt-get install -y curl git unzip xz-utils zip libglu1-mesa
RUN git clone https://github.com/flutter/flutter.git /usr/local/flutter
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"
RUN flutter config --enable-web
WORKDIR /app
COPY . .
RUN flutter clean
RUN flutter pub get
` + build + `

FROM scratch
COPY --from=0 /app/build/web /
`

		// We run it with caddy for Containerized mode.
		if meta["serverless"] != "true" {
			caddy := `
FROM zeabur/caddy-static AS runtime
COPY --from=1 / /usr/share/caddy
`

			dockerfile += caddy
		}

		return dockerfile, nil
	}

	if meta["framework"] == "serverpod" {
		return `FROM dart:3.2.5 AS build
WORKDIR /app
COPY . .
RUN dart pub get
` + build + `
CMD ["/app/bin/main", "--apply-migrations"]
`, nil
	}

	return `FROM dart:stable-sdk
RUN dart pub get
` + build + `

FROM alpine:latest
COPY --from=0 /app/bin/main /main
CMD ["/main"]
`, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Node.js packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
