package nodejs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestIsMpaFramework_WithoutFramework(t *testing.T) {
	if isMpaFramework("") {
		t.Error("should return false")
	}
}

func TestIsNotMpaFramework_WithoutFramework(t *testing.T) {
	if !isNotMpaFramework("") {
		t.Error("should return true")
	}
}

func TestIsMpaFramework_DefaultFalse(t *testing.T) {
	if isMpaFramework("aaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Error("should return false")
	}
}

func TestIsMpaFramework_CanTrue(t *testing.T) {
	if isNotMpaFramework("hexo") {
		t.Error("should return true")
	}
}

// TODO)) type-safe builder
func TestGetContextBasedOnMeta_MapShouldBeCorrect(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
		SPA:         true,
	})
}

func TestGetContextBasedOnMeta_WithOutputdirAndSPAFramework(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
		"outputDir":   "dist",
		"framework":   "wtfisthis",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
		OutputDir:   "dist",
		SPA:         true,
	})
}

func TestGetContextBasedOnMeta_WithOutputdirAndMPAFramework(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
		"outputDir":   "dist",
		"framework":   "hexo",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
		OutputDir:   "dist",
		SPA:         false,
	})
}
