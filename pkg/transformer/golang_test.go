package transformer_test

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/transformer"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestTransformGolang(t *testing.T) {
	t.Parallel()

	t.Run("not a go project", func(t *testing.T) {
		t.Parallel()

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeDeno,
			PlanMeta:     map[string]string{},
			BuildkitPath: afero.NewMemMapFs(),
			AppPath:      afero.NewMemMapFs(),
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformGolang(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("not a serverless project", func(t *testing.T) {
		t.Parallel()

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeGo,
			PlanMeta:     map[string]string{},
			BuildkitPath: afero.NewMemMapFs(),
			AppPath:      afero.NewMemMapFs(),
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformGolang(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("serverless", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		_ = afero.WriteFile(buildkitPath, "main", []byte("data"), 0o644)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeGo,
			PlanMeta:     map[string]string{"serverless": "true"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformGolang(ctx)
		require.NoError(t, err)

		// create snapshot
		SnapshotFs(t, "golang-serverless", ctx.AppPath)
	})
}
