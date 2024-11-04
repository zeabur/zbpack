package transformer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/transformer"
	"github.com/zeabur/zbpack/pkg/types"
)

// TestTransformRust transforms Rust functions.
func TestTransformRust(t *testing.T) {
	t.Parallel()

	t.Run("not a rust project", func(t *testing.T) {
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

		err := transformer.TransformRust(ctx)
		require.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("not a serverless project", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "empty")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeRust,
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformRust(ctx)
		require.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("serverless", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "binary-serverless")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypeRust,
			PlanMeta:     map[string]string{"serverless": "true"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformRust(ctx)
		require.NoError(t, err)
	})
}
