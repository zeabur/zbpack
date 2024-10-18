package zeaburpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectDockerfile(t *testing.T) {
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

	injectedDockerfile := InjectDockerfile(dockerfile, &registry, variables)

	expectedDockerfile := `FROM test.io/library/alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""


RUN echo hello

FROM test.io/library/alpine:3.12 AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""


RUN echo world`

	assert.Equal(t, injectedDockerfile, expectedDockerfile)

	injectedDockerfileWithoutRegistry := InjectDockerfile(dockerfile, nil, variables)

	expectedDockerfileWithoutRegistry := `FROM alpine:3.12 AS builder
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""


RUN echo hello

FROM alpine:3.12 AS runner
ENV KEY="VALUE"
ENV KEY2="\"Value\\\"\""


RUN echo world`

	assert.Equal(t, injectedDockerfileWithoutRegistry, expectedDockerfileWithoutRegistry)
}
