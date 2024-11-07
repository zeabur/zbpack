package transformer

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestTransformStatic(t *testing.T) {
	t.Run("should skip when no output dir and not serverless static", func(t *testing.T) {
		ctx := &Context{
			PlanMeta: map[string]string{
				"outputDir": "",
			},
			PlanType: types.PlanTypeStatic,
		}

		err := TransformStatic(ctx)
		assert.ErrorIs(t, err, ErrSkip)
	})

	t.Run("should not skip if there is serverless", func(t *testing.T) {
		ctx := &Context{
			PlanMeta: map[string]string{
				"serverless": "true",
			},
			PlanType: types.PlanTypeStatic,
		}

		err := TransformStatic(ctx)
		assert.NotErrorIs(t, err, ErrSkip)
	})

	t.Run("should copy files and create config for SPA", func(t *testing.T) {
		// Setup temp dirs
		tempDir := t.TempDir()
		buildkitPath := path.Join(tempDir, "buildkit")
		appPath := path.Join(tempDir, "app")

		err := os.MkdirAll(buildkitPath, 0o755)
		assert.NoError(t, err)

		err = os.WriteFile(path.Join(buildkitPath, "index.html"), []byte("test"), 0o644)
		assert.NoError(t, err)

		ctx := &Context{
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PlanMeta: map[string]string{
				"outputDir":  "dist",
				"serverless": "true",
			},
			PlanType: types.PlanTypeStatic,
		}

		err = TransformStatic(ctx)
		assert.NoError(t, err)

		// Verify files copied
		_, err = os.Stat(path.Join(appPath, ".zeabur/output/static/index.html"))
		assert.NoError(t, err)

		// Verify config.json
		configBytes, err := os.ReadFile(path.Join(appPath, ".zeabur/output/config.json"))
		assert.NoError(t, err)

		var config types.ZeaburOutputConfig
		err = json.Unmarshal(configBytes, &config)
		assert.NoError(t, err)
		assert.Len(t, config.Routes, 1)
		assert.Equal(t, ".*", config.Routes[0].Src)
		assert.Equal(t, "/", config.Routes[0].Dest)
	})

	t.Run("should handle hidden files correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		buildkitPath := path.Join(tempDir, "buildkit")
		appPath := path.Join(tempDir, "app")

		err := os.MkdirAll(buildkitPath, 0o755)
		assert.NoError(t, err)

		// Create some hidden files/dirs
		err = os.WriteFile(path.Join(buildkitPath, ".hidden"), []byte("test"), 0o644)
		assert.NoError(t, err)

		err = os.MkdirAll(path.Join(buildkitPath, ".hiddendir"), 0o755)
		assert.NoError(t, err)

		// Create .well-known dir which should be preserved
		err = os.MkdirAll(path.Join(buildkitPath, ".well-known"), 0o755)
		assert.NoError(t, err)

		ctx := &Context{
			BuildkitPath: buildkitPath,
			AppPath:      appPath,
			PlanMeta: map[string]string{
				"outputDir":  "dist",
				"serverless": "true",
			},
			PlanType: types.PlanTypeStatic,
		}

		err = TransformStatic(ctx)
		assert.NoError(t, err)

		// Verify hidden files/dirs removed
		_, err = os.Stat(path.Join(appPath, ".zeabur/output/static/.hidden"))
		assert.True(t, os.IsNotExist(err))

		_, err = os.Stat(path.Join(appPath, ".zeabur/output/static/.hiddendir"))
		assert.True(t, os.IsNotExist(err))

		// Verify .well-known preserved
		_, err = os.Stat(path.Join(appPath, ".zeabur/output/static/.well-known"))
		assert.NoError(t, err)
	})
}
