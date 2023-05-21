package dockerfile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFindDockerfile_WithUppercase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", path)
}

func TestFindDockerfile_WithLowercase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dockerfile", path)
}

func TestFindDockerfile_WithRandomcase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dOckErFIle", []byte("FROM alpine"), 0o644)

	ctx := dockerfilePlanContext{
		src: fs,
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dOckErFIle", path)
}

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

func TestGetMeta_Content(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	meta := GetMeta(GetMetaOptions{Src: fs})

	assert.Equal(t, "FROM alpine", meta["content"])
}
