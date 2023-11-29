package nodejs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

// TODO)) type-safe builder
func TestGetContextBasedOnMeta_MapShouldBeCorrect(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "RUN npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "RUN npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
	})
}

func TestGetContextBasedOnMeta_WithOutputdirAndSPAFramework(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "RUN npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
		"outputDir":   "dist",
		"framework":   "wtfisthis",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "RUN npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
		Framework:   "wtfisthis",
		OutputDir:   "dist",
	})
}

func TestGetContextBasedOnMeta_WithOutputdirAndMPAFramework(t *testing.T) {
	meta := getContextBasedOnMeta(types.PlanMeta{
		"nodeVersion": "16",
		"installCmd":  "RUN npm install",
		"buildCmd":    "npm run build",
		"startCmd":    "npm run start",
		"outputDir":   "dist",
		"framework":   "hexo",
	})

	assert.Equal(t, meta, TemplateContext{
		NodeVersion: "16",
		InstallCmd:  "RUN npm install",
		BuildCmd:    "npm run build",
		StartCmd:    "npm run start",
		Framework:   "hexo",
		OutputDir:   "dist",
	})
}
