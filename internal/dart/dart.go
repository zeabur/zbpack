// Package dart provides a Dart packer.
package dart

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Dart projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	if meta["framework"] == "flutter" {
		return `FROM zeabur/flutter
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

FROM busybox:1.36.1-glibc

COPY --from=build /runtime/ /
COPY --from=build /app/bin/main /app/bin/main
COPY --from=build /app/config/ config/
COPY --from=build /app/web/ web/

CMD app/bin/main
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
