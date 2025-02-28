package dart_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/dart"
)

func TestGenerateDockerfileStatic(t *testing.T) {
	t.Parallel()

	dockerfile, err := dart.GenerateDockerfile(map[string]string{
		"framework": "flutter",
	})

	assert.NoError(t, err)
	assert.Contains(t, dockerfile, "FROM scratch")
	assert.Contains(t, dockerfile, "FROM zeabur/caddy-static")
}
