// Package swift is the build planner for Swift projects.
package swift

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

type pack struct {
	*identify
}

// NewPacker returns a new Packer for Swift
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

// GenerateDockerfile generates a Dockerfile for Swift project
func GenerateDockerfile(_ types.PlanMeta) (string, error) {
	// TODO: following dockerfile is copied from Vapor's template, need to be modified to support other Swift use cases
	return `FROM swift:5.9-jammy AS build

RUN export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true \
  && apt-get -q update \
  && apt-get -q dist-upgrade -y \
  && apt-get install -y libjemalloc-dev

WORKDIR /build

COPY ./Package.* ./
RUN swift package resolve --skip-update \
  $([ -f ./Package.resolved ] && echo "--force-resolved-versions" || true)

COPY . .

RUN swift build -c release \
  --static-swift-stdlib \
  -Xlinker -ljemalloc

WORKDIR /staging

RUN cp "$(swift build --package-path /build -c release --show-bin-path)/App" ./

RUN cp "/usr/libexec/swift/linux/swift-backtrace-static" ./

RUN find -L "$(swift build --package-path /build -c release --show-bin-path)/" -regex '.*\.resources$' -exec cp -Ra {} ./ \;

RUN [ -d /build/Public ] && { mv /build/Public ./Public && chmod -R a-w ./Public; } || true
RUN [ -d /build/Resources ] && { mv /build/Resources ./Resources && chmod -R a-w ./Resources; } || true

FROM ubuntu:jammy

RUN export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true \
  && apt-get -q update \
  && apt-get -q dist-upgrade -y \
  && apt-get -q install -y \
  libjemalloc2 \
  ca-certificates \
  tzdata \
  libcurl4 \
  libxml2 \
  && rm -r /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build --chown=vapor:vapor /staging /app

ENV SWIFT_BACKTRACE=enable=yes,sanitize=yes,threads=all,images=all,interactive=no,swift-backtrace=./swift-backtrace-static

ENTRYPOINT ["./App"]
CMD ["serve", "--env", "production", "--hostname", "0.0.0.0", "--port", "8080"]
`, nil
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
