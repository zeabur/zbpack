package dockerfile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestFindDockerfile_WithUppercase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", path)
}

func TestFindDockerfile_WithLowercase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dockerfile", path)
}

func TestFindDockerfile_WithRandomcase(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dOckErFIle", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dOckErFIle", path)
}

func TestFindDockerfile_WithSubmodule(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "Dockerfile.Subm", []byte("FROM ubuntu"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile.Subm", path)
}

func TestFindDockerfile_CaseInsensitiveSubmodule(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dOckErFIle", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "dOckErFIle.SUbM", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dOckErFIle.SUbM", path)
}

func TestFindDockerfile_WithSubmodulePrefixed(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "Subm.Dockerfile", []byte("FROM ubuntu"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Subm.Dockerfile", path)
}

func TestFindDockerfile_PrefixedCaseInsensitiveSubmodule(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dOckErFIle", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "sUbM.dOckErFIle", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm",
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "sUbM.dOckErFIle", path)
}

func TestFindDockerfile_NoSuchSubmodule(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dOckErFIle", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "dOckErFIle.SUbM", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source:        fs,
			Config:        config,
			SubmoduleName: "Subm2",
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "dOckErFIle", path)
}

func TestFindDockerfile_WithConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)
	_ = afero.WriteFile(fs, "Dockerfile.test", []byte("FROM ubuntu"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")
	config.Set("dockerfile.name", "test")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	path, err := FindDockerfile(&ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile.test", path)
}

func TestGetExposePort_WithExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine\nEXPOSE 1145"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithoutExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "8080", port)
}

func TestGetExposePort_WithLowercaseExposeSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine\nexpose 1145"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithLowercaseDockerfileSpecified(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine\nEXPOSE 1145"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithSpaceAfterExpose(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine\nEXPOSE 1145 "), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetExposePort_WithLowercaseExpose(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "dockerfile", []byte("FROM alpine\nexpose 1145"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "Subm")

	ctx := dockerfilePlanContext{
		NewPlannerOptions: plan.NewPlannerOptions{
			Source: fs,
			Config: config,
		},
	}
	port := GetExposePort(&ctx)

	assert.Equal(t, "1145", port)
}

func TestGetMeta_Content(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "Dockerfile", []byte("FROM alpine"), 0o644)

	config := plan.NewProjectConfigurationFromFs(fs, "")

	meta := GetMeta(plan.NewPlannerOptions{Source: fs, Config: config})

	assert.Equal(t, "FROM alpine", meta["content"])
}
