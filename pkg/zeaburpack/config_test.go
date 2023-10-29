package zeaburpack_test

import (
	"testing"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

func TestUpdateOptionsOnConfig_FromConfig(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{
		"build_command": "build",
		"start_command": "start",
		"output_dir": "dist"
	}`), 0644)

	options := &zeaburpack.PlanOptions{}
	config := plan.NewProjectConfigurationFromFs(fs, "")

	zeaburpack.UpdateOptionsOnConfig(options, config)

	assert.Equal(t, "build", *options.CustomBuildCommand)
	assert.Equal(t, "start", *options.CustomStartCommand)
	assert.Equal(t, "dist", *options.OutputDir)
}

func TestUpdateOptionsOnConfig_NotUpdateIfDefined(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{
		"build_command": "build",
		"start_command": "start",
		"output_dir": "dist"
	}`), 0644)

	options := &zeaburpack.PlanOptions{
		CustomBuildCommand: lo.ToPtr("build2"),
	}
	config := plan.NewProjectConfigurationFromFs(fs, "")

	zeaburpack.UpdateOptionsOnConfig(options, config)

	assert.Equal(t, "build2", *options.CustomBuildCommand)
	assert.Equal(t, "start", *options.CustomStartCommand)
	assert.Equal(t, "dist", *options.OutputDir)
}

func TestUpdateOptionsOnConfig_LeaveIfAllEmpty(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{
		"start_command": "start",
		"output_dir": "dist"
	}`), 0644)

	options := &zeaburpack.PlanOptions{}
	config := plan.NewProjectConfigurationFromFs(fs, "")

	zeaburpack.UpdateOptionsOnConfig(options, config)

	assert.Nil(t, options.CustomBuildCommand)
	assert.Equal(t, "start", *options.CustomStartCommand)
	assert.Equal(t, "dist", *options.OutputDir)
}

func TestUpdateOptionsOnConfig_CheckType(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "zbpack.json", []byte(`{
		"build_command": 123,
		"start_command": 123,
		"output_dir": 123
	}`), 0644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	t.Run("PlanOptions", func(t *testing.T) {
		t.Parallel()

		options := &zeaburpack.PlanOptions{}
		zeaburpack.UpdateOptionsOnConfig(options, config)

		assert.Equal(t, "123", *options.CustomBuildCommand)
		assert.Equal(t, "123", *options.CustomStartCommand)
		assert.Equal(t, "123", *options.OutputDir)
	})

	t.Run("BuildOptions", func(t *testing.T) {
		t.Parallel()

		options := &zeaburpack.BuildOptions{}
		zeaburpack.UpdateOptionsOnConfig(options, config)

		assert.Equal(t, "123", *options.CustomBuildCommand)
		assert.Equal(t, "123", *options.CustomStartCommand)
		assert.Equal(t, "123", *options.OutputDir)
	})
}
