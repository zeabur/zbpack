package rust

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNeedOpenssl_CargoLockfile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "Cargo.lock", []byte("openssl"), 0o644)

	assert.True(t, needOpenssl(fs))
}

func TestNeedOpenssl_CargoTomlfile(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "Cargo.toml", []byte("openssl"), 0o644)

	assert.True(t, needOpenssl(fs))
}

func TestNeedOpenssl_None(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "Cargo.toml", []byte(""), 0o644)

	assert.False(t, needOpenssl(fs))
}
