package dotnet

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
)

func TestMatch_NotFound(t *testing.T) {
	identifier := NewIdentifier()

	fs := afero.NewMemMapFs()

	assert.False(t, identifier.Match(fs))
}

func TestMatch_Found(t *testing.T) {
	path := "../../tests/dotnet-samples/dotnetapp/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	identifier := NewIdentifier()
	assert.True(t, identifier.Match(fs))
}

func TestPlanMeta_NotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	options := plan.NewPlannerOptions{
		Source:        fs,
		SubmoduleName: "",
	}

	var err error
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	identifier := NewIdentifier()
	identifier.PlanMeta(options)

	if err != nil {
		t.Errorf("Expected panic with message 'Unable to determine SDK version', got %v", err)
	}
}

func TestPlanMeta_Found(t *testing.T) {
	path := "../../tests/dotnet-samples/dotnetapp/"
	assert.DirExists(t, path)

	fs := afero.NewBasePathFs(afero.NewOsFs(), path)

	options := plan.NewPlannerOptions{
		Source:        fs,
		SubmoduleName: "dotnetapp",
	}

	identifier := NewIdentifier()
	planMeta := identifier.PlanMeta(options)

	assert.NotEmpty(t, planMeta)
	assert.Equal(t, "7.0", planMeta["sdk"])
	assert.Equal(t, "dotnetapp.csproj", planMeta["entryPoint"])
}

func TestPlanMeta_NoCsproj(t *testing.T) {
	fs := afero.NewMemMapFs()
	identifier := NewIdentifier()

	planMeta := identifier.PlanMeta(plan.NewPlannerOptions{
		Source:        fs,
		SubmoduleName: "test",
	})

	assert.Equal(t, plan.Continue(), planMeta)
}

func TestPlanMeta_MultipleCsproj(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "test.csproj", []byte(`<Project Sdk="Microsoft.NET.Sdk" ToolsVersion="15.0">

	<PropertyGroup>
	  <OutputType>Exe</OutputType>
	  <TargetFramework>net8.0</TargetFramework>
	  <Nullable>enable</Nullable>
	  <PublishRelease>true</PublishRelease>
	</PropertyGroup>

  </Project>`), 0o644)
	_ = afero.WriteFile(fs, "test2.csproj", []byte(`<Project Sdk="Microsoft.NET.Sdk" ToolsVersion="15.0">

	<PropertyGroup>
	  <OutputType>Exe</OutputType>
	  <TargetFramework>net7.0</TargetFramework>
	  <Nullable>enable</Nullable>
	  <PublishRelease>true</PublishRelease>
	</PropertyGroup>

  </Project>`), 0o644)

	t.Run("project", func(t *testing.T) {
		t.Parallel()

		identifier := NewIdentifier()
		planMeta := identifier.PlanMeta(plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "test",
		})

		assert.Equal(t, "8.0", planMeta["sdk"])
		assert.Equal(t, "test.csproj", planMeta["entryPoint"])
	})

	t.Run("project2", func(t *testing.T) {
		t.Parallel()

		identifier := NewIdentifier()
		planMeta := identifier.PlanMeta(plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "test2",
		})

		assert.Equal(t, "7.0", planMeta["sdk"])
		assert.Equal(t, "test2.csproj", planMeta["entryPoint"])
	})
}
