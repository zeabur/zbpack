package elixir

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDetermineElixirVersion_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()

	ver, err := DetermineElixirVersion(fs)
	assert.ErrorContains(t, err, "unable to determine Elixir version")
	assert.Empty(t, ver)
}

func TestDetermineElixirVersion_Valid(t *testing.T) {
	path := "../../tests/elixir-cases/elixir/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	ver, err := DetermineElixirVersion(fs)
	assert.NoError(t, err)
	assert.Equal(t, ver, "1.12")
}

func TestDetermineElixirFramework_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()

	framework, err := DetermineElixirFramework(fs)
	assert.ErrorContains(t, err, "unable to determine Elixir framework")
	assert.Empty(t, framework)
}

func TestDetermineElixirFramework_Valid(t *testing.T) {
	path := "../../tests/elixir-cases/elixir/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	framework, err := DetermineElixirFramework(fs)
	assert.NoError(t, err)
	assert.Equal(t, framework, "phoenix")
}

func TestDetermineElixirFramework_Invalid(t *testing.T) {
	path := "../../tests/elixir-cases/elixir_ecto/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	framework, err := DetermineElixirFramework(fs)
	assert.NoError(t, err)
	assert.Equal(t, framework, "")
}

func TestCheckElixirEcto_Empty(t *testing.T) {
	fs := afero.NewMemMapFs()

	usesEcto, err := CheckElixirEcto(fs)
	assert.ErrorContains(t, err, "unable to determine if Ecto is used")
	assert.Empty(t, usesEcto)
}

func TestCheckElixirEcto_Valid(t *testing.T) {
	path := "../../tests/elixir-cases/elixir_ecto/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	usesEcto, err := CheckElixirEcto(fs)
	assert.NoError(t, err)
	assert.Equal(t, usesEcto, "true")
}

func TestCheckElixirEcto_Invalid(t *testing.T) {
	path := "../../tests/elixir-cases/elixir/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	usesEcto, err := CheckElixirEcto(fs)
	assert.NoError(t, err)
	assert.Equal(t, usesEcto, "false")
}
