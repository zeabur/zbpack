package nodejs_test

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/nodejs"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run `go-snaps` will sort snapshots
	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestTemplate_NBuildCmd_NOutputDir(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InstallCmd: "RUN yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_NBuildCmd_OutputDir_NSPA(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InstallCmd: "RUN yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_NBuildCmd_OutputDir_SPA(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InstallCmd: "RUN yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_NOutputDir(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InstallCmd: "RUN yarn install",
		BuildCmd:   "RUN yarn build",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_OutputDir(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InstallCmd: "RUN yarn install",
		BuildCmd:   "RUN yarn build",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_Bun(t *testing.T) {
	ctx := nodejs.TemplateContext{
		Bun:         true,
		NodeVersion: "18",
		InstallCmd:  "RUN bun install",
		StartCmd:    "bun start main.ts",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}
