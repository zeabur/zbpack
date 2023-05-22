package dotnet

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDetermineSDKVersion_EmptyEntryPoint(t *testing.T) {
	fs := afero.NewMemMapFs()

	ver, err := DetermineSDKVersion("", fs)
	assert.ErrorContains(t, err, "Unable to determine SDK version")
	assert.Empty(t, ver)
}

func TestDetermineSDKVersion_Valid(t *testing.T) {
	path := "../../tests/dotnet-samples/dotnetapp/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	ver, err := DetermineSDKVersion("dotnetapp", fs)
	assert.NoError(t, err)
	assert.Equal(t, ver, "7.0")
}
