package golang_test

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/internal/golang"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestGenerateDockerfile_CGO(t *testing.T) {
	t.Parallel()

	baseMeta := types.PlanMeta{
		"goVersion": "1.22",
		"entry":     "main.go",
	}

	t.Run("CGO enabled", func(t *testing.T) {
		t.Parallel()

		meta := maps.Clone(baseMeta)
		meta["cgo"] = "true"

		dockerfile, err := golang.GenerateDockerfile(meta)
		require.NoError(t, err)

		assert.Contains(t, dockerfile, "ENV CGO_ENABLED=1\n")
		assert.Contains(t, dockerfile, "RUN apk add --no-cache build-base cmake\n")
	})

	t.Run("CGO disabled", func(t *testing.T) {
		t.Parallel()

		meta := maps.Clone(baseMeta)
		meta["cgo"] = "false"

		dockerfile, err := golang.GenerateDockerfile(meta)
		require.NoError(t, err)

		assert.Contains(t, dockerfile, "ENV CGO_ENABLED=0\n")
		assert.NotContains(t, dockerfile, "RUN apk add --no-cache build-base cmake\n")
	})
}

func TestGenerateDockerfile_BuildCommand(t *testing.T) {
	t.Parallel()

	baseMeta := types.PlanMeta{
		"goVersion": "1.22",
		"entry":     "main.go",
	}

	t.Run("with build command", func(t *testing.T) {
		t.Parallel()

		meta := maps.Clone(baseMeta)
		meta["buildCommand"] = "go generate ./..."

		dockerfile, err := golang.GenerateDockerfile(meta)
		require.NoError(t, err)

		assert.Contains(t, dockerfile, "RUN go generate ./...\n\nRUN go build -o ./bin/server")
	})

	t.Run("without build command", func(t *testing.T) {
		t.Parallel()

		meta := maps.Clone(baseMeta)

		dockerfile, err := golang.GenerateDockerfile(meta)
		require.NoError(t, err)

		assert.NotContains(t, dockerfile, "RUN go generate ./...\n")
		assert.Contains(t, dockerfile, "RUN go build -o ./bin/server")
	})

	t.Run("cgo + build command", func(t *testing.T) {
		t.Parallel()

		meta := maps.Clone(baseMeta)
		meta["cgo"] = "true"
		meta["buildCommand"] = "go generate ./..."

		dockerfile, err := golang.GenerateDockerfile(meta)
		require.NoError(t, err)

		assert.Contains(t, dockerfile, "ENV CGO_ENABLED=1\nRUN go generate ./...\n\nRUN go build -o ./bin/server")
	})
}
