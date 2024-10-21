package transformer_test

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/transformer"
)

func TestTransformerZeaburDir(t *testing.T) {
	t.Parallel()

	t.Run("has .zeabur directory", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()
		_ = afero.WriteFile(buildkitPath, ".zeabur/aaa", []byte("Hello"), 0o644)
		_ = afero.WriteFile(buildkitPath, ".zeabur/bbb/text.txt", []byte("World"), 0o644)

		context := &transformer.Context{
			PlanType:     "",
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformZeaburDir(context)
		assert.NoError(t, err)

		_, err = appPath.Stat(".zeabur")
		assert.NoError(t, err)

		data, err := afero.ReadFile(appPath, ".zeabur/aaa")
		assert.NoError(t, err)
		assert.Equal(t, "Hello", string(data))

		data, err = afero.ReadFile(appPath, ".zeabur/bbb/text.txt")
		assert.NoError(t, err)
		assert.Equal(t, "World", string(data))
	})

	t.Run("no .zeabur directory", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		context := &transformer.Context{
			PlanType:     "",
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformZeaburDir(context)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})
}
