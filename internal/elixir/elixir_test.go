package elixir

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestGenerateDockerFile_Valid(t *testing.T) {
	planMeta := types.PlanMeta{
		"ver":       "1.15",
		"framework": "phoenix",
		"ecto":      "false",
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
			s:        "elixir:1.15",
			expected: true,
		},
		{
			s:        "mix assets.deploy",
			expected: true,
		},
		{
			s:        "mix phx.server",
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
