package dockerfile_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/dockerfile"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	identifier := dockerfile.NewIdentifier()
	assert.True(t, identifier.Match(fs))
}

func TestMatch_SuffixDockerfile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile.a", []byte("FROM alpine"), 0o644)

	identifier := dockerfile.NewIdentifier()
	assert.True(t, identifier.Match(fs))
}

func TestMatch_PrefixDockerfile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "a.Dockerfile", []byte("FROM alpine"), 0o644)

	identifier := dockerfile.NewIdentifier()
	assert.True(t, identifier.Match(fs))
}
