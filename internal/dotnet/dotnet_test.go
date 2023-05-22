package dotnet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestGenerateDockerFile_Valid(t *testing.T) {
	planMeta := types.PlanMeta{
		"sdk":        "7.0",
		"entryPoint": "dotnetapp",
	}

	dockerfile, err := GenerateDockerfile(planMeta)
	assert.Empty(t, err)

	tests := []struct {
		s        string
		expected bool
	}{
		{
			s:        "",
			expected: false,
		},
		{
			s:        "sdk:7.0",
			expected: true,
		},
		{
			s:        "dotnetapp.dll",
			expected: true,
		},
	}

	contains := func(dockerfile, s string) bool {
		if s == "" {
			return false
		}

		return strings.Contains(dockerfile, s)
	}

	for _, test := range tests {
		assert.Equal(t, contains(dockerfile, test.s), test.expected)
	}
}
