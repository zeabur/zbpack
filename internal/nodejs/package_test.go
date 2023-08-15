package nodejs_test

import (
	"encoding/json"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/nodejs"
)

func TestNewPackageJson(t *testing.T) {
	p := nodejs.NewPackageJSON()

	// dependencies
	assert.NotPanics(t, func() {
		_, ok := p.Dependencies["astro"]
		assert.False(t, ok)
	})

	// devDependencies
	assert.NotPanics(t, func() {
		_, ok := p.DevDependencies["astro"]
		assert.False(t, ok)
	})

	// scripts
	assert.NotPanics(t, func() {
		_, ok := p.Scripts["build"]
		assert.False(t, ok)
	})

	// engines
	assert.Empty(t, p.Engines.Node)

	// main
	assert.Empty(t, p.Main)
}

func TestDeserializePackageJson_NoSuchFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.Error(t, err)
}

func TestDeserializePackageJson_InvalidFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte("\\0"), 0o644)

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.Error(t, err)
}

func TestDeserializePackageJson_EmptyJson(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte("{}"), 0o644)

	p, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)
	assert.Empty(t, p.Dependencies)
	assert.Empty(t, p.DevDependencies)
	assert.Empty(t, p.Scripts)
	assert.Empty(t, p.Engines.Node)
	assert.Empty(t, p.Main)
}

func TestDeserializePackageJson_NoMatchField(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{"this_is_a_test": "hhh"}`), 0o644)

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)
}

func TestDeserializePackageJson_WithDepsAndEngines(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"dependencies": map[string]string{
			"astro": "0.0.1",
		},
		"devDependencies": map[string]string{
			"prettier": "^1.2.3",
		},
		"engines": map[string]string{
			"node": "^18",
		},
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	version, ok := packageJSON.Dependencies["astro"]
	assert.True(t, ok)
	assert.Equal(t, "0.0.1", version)

	version, ok = packageJSON.DevDependencies["prettier"]
	assert.True(t, ok)
	assert.Equal(t, "^1.2.3", version)

	assert.Equal(t, "^18", packageJSON.Engines.Node)
}

func TestDeserializePackageJson_WithMainAndEngines(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"main": "hello",
		"engines": map[string]string{
			"node": "^18",
		},
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.Equal(t, "hello", packageJSON.Main)
	assert.Equal(t, "^18", packageJSON.Engines.Node)
}

func TestDeserializePackageJson_WithPackageManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"packageManager": "yarn@1.2.3",
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.NotNil(t, packageJSON.PackageManager)
	assert.Equal(t, "yarn@1.2.3", *packageJSON.PackageManager)
}

func TestDeserializePackageJson_WithoutPackageManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.Nil(t, packageJSON.PackageManager)
}

func TestContainsDependency(t *testing.T) {
	p := nodejs.NewPackageJSON()
	p.Dependencies["solid"] = "0.0.1"

	d, ok := p.FindDependency("solid")
	assert.True(t, ok)
	assert.Equal(t, "0.0.1", d)
}

func TestContainsDependency_Dev(t *testing.T) {
	p := nodejs.NewPackageJSON()
	p.DevDependencies["solid-start-node"] = "0.0.2"

	d, ok := p.FindDependency("solid-start-node")
	assert.True(t, ok)
	assert.Equal(t, "0.0.2", d)
}

func TestContainsDependency_NotFound(t *testing.T) {
	p := nodejs.NewPackageJSON()
	p.Dependencies["solid"] = "0.0.1"

	_, ok := p.FindDependency("astro")
	assert.False(t, ok)
}

func TestContainsDependency_Both(t *testing.T) {
	p := nodejs.NewPackageJSON()
	p.Dependencies["solid"] = "0.0.1"
	p.DevDependencies["solid-start-node"] = "0.0.2"

	d, ok := p.FindDependency("solid")
	assert.True(t, ok)
	assert.Equal(t, "0.0.1", d)

	d, ok = p.FindDependency("solid-start-node")
	assert.True(t, ok)
	assert.Equal(t, "0.0.2", d)
}

func TestContainsDependency_Conflict(t *testing.T) {
	p := nodejs.NewPackageJSON()
	p.Dependencies["solid"] = "0.0.1"
	p.DevDependencies["solid"] = "0.0.2"

	// not stable behavior
	d, ok := p.FindDependency("solid")
	assert.True(t, ok)
	assert.Equal(t, "0.0.1", d)
}
