package transformer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/pkg/transformer"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestTransformPython(t *testing.T) {
	t.Parallel()

	t.Run("not a python project", func(t *testing.T) {
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

		err := transformer.TransformPython(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("not a serverless project", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "empty")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformPython(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("serverless-default-entry", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "python-serverless")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{"serverless": "true", "pythonVersion": "3.10"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}
		err := transformer.TransformPython(ctx)
		require.NoError(t, err)
	})

	t.Run("serverless-custom-entry", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "python-serverless-custom-entry")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{"serverless": "true", "pythonVersion": "3.10", "entry": "myentry.py"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformPython(ctx)
		require.NoError(t, err)
	})

	t.Run("serverless-with-static", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "python-serverless-with-static")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{"serverless": "true", "pythonVersion": "3.10"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformPython(ctx)
		require.NoError(t, err)
	})

	t.Run("serverless-with-venv", func(t *testing.T) {
		t.Parallel()

		buildkitPath := GetInputPath(t, "python-serverless-with-venv")
		appPath := GetOutputSnapshotPath(t)

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{"serverless": "true", "pythonVersion": "3.10"},
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformPython(ctx)
		require.NoError(t, err)
	})
}
