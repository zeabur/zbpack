package transformer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/transformer"
)

func TestTransformerZeaburDir(t *testing.T) {
	t.Parallel()

	t.Run("has .zeabur directory", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "zeabur-dir")
		appPath := GetOutputSnapshotPath(t)

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
		require.NoError(t, err)

		stat, err := os.Stat(filepath.Join(buildkitPath, ".zeabur"))
		require.NoError(t, err)
		assert.True(t, stat.IsDir())

		data, err := os.Stat(filepath.Join(buildkitPath, ".zeabur/aaa"))
		require.NoError(t, err)
		assert.False(t, data.IsDir())

		stat, err = os.Stat(filepath.Join(appPath, ".zeabur/bbb"))
		require.NoError(t, err)
		assert.True(t, stat.IsDir())

		data, err = os.Stat(filepath.Join(appPath, ".zeabur/bbb/bbb"))
		require.NoError(t, err)
		assert.False(t, data.IsDir())
	})

	t.Run("no .zeabur directory", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "empty")
		appPath := GetOutputSnapshotPath(t)

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
