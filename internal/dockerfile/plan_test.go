package dockerfile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGetExposePort_WithExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine\nEXPOSE 1145"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithoutExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "8080", port)
}

func TestGetExposePort_WithLowercaseExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine\nexpose 1145"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithLowercaseDockerfileSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine\nEXPOSE 1145"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}
