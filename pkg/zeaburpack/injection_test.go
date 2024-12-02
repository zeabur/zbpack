package zeaburpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestInjectDockerfile(t *testing.T) {
	t.Parallel()

	t.Run("registry", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 AS builder
RUN echo hello

FROM alpine:3.12 AS runner
RUN echo world`

		registry := "test.io"
		variables := map[string]string{
			"KEY":  "VALUE",
			"KEY2": `"Value\""`,
		}

		injectedDockerfile := InjectDockerfile(dockerfile, &registry, variables, types.PlanTypeSwift, types.PlanMeta{
			"framework": "vapor",
		})

		expectedDockerfile := `FROM test.io/library/alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo hello

FROM test.io/library/alpine:3.12 AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo world`

		assert.Equal(t, injectedDockerfile, expectedDockerfile)
	})

	t.Run("without registry", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 AS builder
RUN echo hello

FROM alpine:3.12 AS runner
RUN echo world`

		variables := map[string]string{
			"KEY":  "VALUE",
			"KEY2": `"Value\""`,
		}

		injectedDockerfileWithoutRegistry := InjectDockerfile(dockerfile, nil, variables, types.PlanTypeSwift, types.PlanMeta{
			"framework": "vapor",
		})

		expectedDockerfileWithoutRegistry := `FROM alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo hello

FROM alpine:3.12 AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo world`

		assert.Equal(t, injectedDockerfileWithoutRegistry, expectedDockerfileWithoutRegistry)
	})

	t.Run("multi-stage build, with registry", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 AS builder
RUN echo hello

FROM builder AS runner
RUN echo world`

		registry := "test.io"
		variables := map[string]string{
			"KEY":  "VALUE",
			"KEY2": `"Value\""`,
		}

		injectedDockerfile := InjectDockerfile(dockerfile, &registry, variables, types.PlanTypeSwift, types.PlanMeta{
			"framework": "vapor",
		})

		expectedDockerfile := `FROM test.io/library/alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo hello

FROM builder AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo world`

		assert.Equal(t, injectedDockerfile, expectedDockerfile)
	})

	t.Run("multi-stage build, without registry", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 AS builder
RUN echo hello

FROM builder AS runner
RUN echo world`

		variables := map[string]string{
			"KEY":  "VALUE",
			"KEY2": `"Value\""`,
		}

		injectedDockerfile := InjectDockerfile(dockerfile, nil, variables, types.PlanTypeSwift, types.PlanMeta{
			"framework": "vapor",
		})

		expectedDockerfile := `FROM alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo hello

FROM builder AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo world`

		assert.Equal(t, injectedDockerfile, expectedDockerfile)
	})

	t.Run("multi-stage build, without registry, 'as' lowercase", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 as builder
RUN echo hello

FROM builder AS runner
RUN echo world`

		variables := map[string]string{
			"KEY":  "VALUE",
			"KEY2": `"Value\""`,
		}

		injectedDockerfile := InjectDockerfile(dockerfile, nil, variables, types.PlanTypeSwift, types.PlanMeta{
			"framework": "vapor",
		})

		expectedDockerfile := `FROM alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo hello

FROM builder AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""

LABEL com.zeabur.zbpack.language="swift" com.zeabur.zbpack.framework="vapor"

RUN echo world`

		assert.Equal(t, injectedDockerfile, expectedDockerfile)
	})

	t.Run("without framework", func(t *testing.T) {
		t.Parallel()

		dockerfile := `FROM alpine:3.12 AS builder
RUN echo hello`

		variables := map[string]string{
			"KEY": "VALUE",
		}

		injectedDockerfile := InjectDockerfile(dockerfile, nil, variables, types.PlanTypeDocker, types.PlanMeta{})

		expectedDockerfile := `FROM alpine:3.12 AS builder
ENV KEY="VALUE"

LABEL com.zeabur.zbpack.language="docker" com.zeabur.zbpack.framework=""

RUN echo hello`

		assert.Equal(t, injectedDockerfile, expectedDockerfile)
	})
}
