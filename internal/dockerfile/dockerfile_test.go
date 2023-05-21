package dockerfile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDockerFile(t *testing.T) {
	const expectedContent = "FROM alpine"

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte(expectedContent), 0o644)

	ctx := GetMeta(GetMetaOptions{Src: fs})
	packer := NewPacker()
	actualContent, err := packer.GenerateDockerfile(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedContent, actualContent)
}
