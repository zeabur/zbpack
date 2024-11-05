package gleam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/gleam"
)

func TestGenerateDockerfile(t *testing.T) {
	t.Parallel()

	t.Run("serverless", func(t *testing.T) {
		t.Parallel()

		dockerfile, err := gleam.GenerateDockerfile(map[string]string{
			"serverless": "true",
		})

		assert.NoError(t, err)
		assert.Contains(t, dockerfile, "\nFROM scratch AS output")
	})

	t.Run("non-serverless", func(t *testing.T) {
		t.Parallel()

		dockerfile, err := gleam.GenerateDockerfile(map[string]string{
			"serverless": "false",
		})

		assert.NoError(t, err)
		assert.Contains(t, dockerfile, "\nWORKDIR /app")
	})
}
