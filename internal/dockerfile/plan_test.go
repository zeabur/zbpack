package dockerfile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestGetMeta_Content(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	meta := GetMeta(plan.NewPlannerOptions{Source: fs, Config: config})

	assert.Equal(t, "FROM alpine", meta["content"])
}
