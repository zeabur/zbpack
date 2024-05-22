// Package dart provides a Dart packer.
package dart

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Dart projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	if meta["framework"] == "flutter" {
		return `FROM ubuntu:latest
RUN apt-get update && apt-get install -y curl git unzip xz-utils zip libglu1-mesa
RUN curl -o /tmp/flutter.tar.xz https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.22.0-stable.tar.xz
RUN tar xf /tmp/flutter.tar.xz -C /usr/local
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"
RUN flutter channel master && flutter upgrade && flutter config --enable-web
RUN flutter doctor
WORKDIR /app
COPY . .
RUN flutter clean
RUN flutter pub get
RUN flutter build web

FROM scratch
COPY --from=0 /app/build/web /
`, nil
	}

	if meta["framework"] == "serverpod" {
		return `FROM dart:3.2.5 AS build
WORKDIR /app
COPY . .
RUN dart pub get
RUN dart compile exe bin/main.dart -o bin/main
CMD ["/app/bin/main", "--apply-migrations"]
`, nil
	}

	return `FROM dart:stable-sdk
RUN dart pub get
RUN dart compile exe bin/main.dart

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
