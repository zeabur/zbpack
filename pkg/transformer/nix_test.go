package transformer_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestNixPart(t *testing.T) {
	t.Parallel()

	t.Run("afero-test-delete-root", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = fs.MkdirAll("/tmp", 0o755)

		tmpFs := afero.NewBasePathFs(fs, "/tmp")
		err := tmpFs.Remove("")
		require.NoError(t, err)

		_, err = fs.Stat("/tmp")
		require.Error(t, err)
	})
}
