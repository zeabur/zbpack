package gleam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/gleam"
)

func TestGenerateDockerfile(t *testing.T) {
	t.Parallel()

	dockerfile, err := gleam.GenerateDockerfile(map[string]string{})

	assert.NoError(t, err)
	assert.Contains(t, dockerfile, "\nWORKDIR /app")
}
