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

func TestTransformPython(t *testing.T) {
	t.Parallel()

	t.Run("not a python project", func(t *testing.T) {
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

		err := transformer.TransformPython(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("not a serverless project", func(t *testing.T) {
		t.Parallel()

		ctx := &transformer.Context{
			PlanType:     types.PlanTypePython,
			PlanMeta:     map[string]string{},
			BuildkitPath: afero.NewMemMapFs(),
			AppPath:      afero.NewMemMapFs(),
			PushImage:    false,
			ResultImage:  "",
			LogWriter:    os.Stderr,
		}

		err := transformer.TransformPython(ctx)
		assert.ErrorIs(t, err, transformer.ErrSkip)
	})

	t.Run("serverless-default-entry", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		_ = afero.WriteFile(buildkitPath, "main.py", []byte("print('hi')"), 0o644)

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

		SnapshotFs(t, "python-serverless-default-entry", appPath)
	})

	t.Run("serverless-custom-entry", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		_ = afero.WriteFile(buildkitPath, "myentry.py", []byte("print('hi')"), 0o644)

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

		SnapshotFs(t, "python-serverless-custom-entry", appPath)
	})

	t.Run("serverless-with-static", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		_ = afero.WriteFile(buildkitPath, "main.py", []byte("print('hi')"), 0o644)
		_ = afero.WriteFile(buildkitPath, "static/index.html", []byte("<html></html>"), 0o644)

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

		SnapshotFs(t, "python-serverless-with-static", appPath)
	})

	t.Run("serverless-with-venv", func(t *testing.T) {
		t.Parallel()

		appPath := afero.NewMemMapFs()
		buildkitPath := afero.NewMemMapFs()

		_ = afero.WriteFile(buildkitPath, "main.py", []byte("print('hi')"), 0o644)
		_ = afero.WriteFile(buildkitPath, "lib/python3.10/site-packages/mypackage/__init__.py", []byte("print('mypackage')"), 0o644)

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

		SnapshotFs(t, "python-serverless-with-venv", appPath)
	})
}
