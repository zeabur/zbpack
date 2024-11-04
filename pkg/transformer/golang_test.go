package transformer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/transformer"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestTransformGolang(t *testing.T) {
	t.Parallel()

	t.Run("not a go project", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "empty")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeDeno,
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformGolang(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("not a serverless project", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "empty")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeGo,
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformGolang(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("serverless", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "binary-serverless")
		appPath := GetOutputSnapshotPath(t)

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
	})
}
