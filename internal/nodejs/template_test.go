package nodejs_test

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		InitCmd:    "RUN npm install -g yarn@latest",
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

		InitCmd:    "RUN npm install -g yarn@latest",
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

		InitCmd:    "RUN npm install -g yarn@latest",
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

		InitCmd:    "RUN npm install -g yarn@latest",
		InstallCmd: "RUN yarn install",
		BuildCmd:   "yarn build",
		StartCmd:   "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_OutputDir(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",

		InitCmd:    "RUN npm install -g yarn@latest",
		InstallCmd: "RUN yarn install",
		BuildCmd:   "yarn build",
		StartCmd:   "yarn start",

		OutputDir: "/app/dist",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_Monorepo(t *testing.T) {
	ctx := nodejs.TemplateContext{
		NodeVersion: "18",
		AppDir:      "myservice",
		InitCmd:     "RUN npm install -g yarn@latest",
		InstallCmd:  "WORKDIR /src/myservice\nRUN yarn install",
		StartCmd:    "yarn start",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_WithOutputDir(t *testing.T) {
	t.Parallel()

	ctx := nodejs.TemplateContext{
		NodeVersion: "18",
		InitCmd:     "RUN npm install -g yarn@latest",
		InstallCmd:  "RUN yarn install",
		OutputDir:   "/app/dist",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)

	require.Contains(t, result, "FROM scratch AS output")
	require.Contains(t, result, "FROM zeabur/caddy-static AS runtime")
}

func TestTemplate_NitroRejectsZeaburOutput(t *testing.T) {
	t.Parallel()

	ctx := nodejs.TemplateContext{
		NodeVersion: "22",
		InitCmd:     "RUN npm install -f -g pnpm@10 || npm install -f -g pnpm@8",
		InstallCmd:  "RUN pnpm install",
		BuildCmd:    "NITRO_PRESET=node-server SERVER_PRESET=node-server pnpm build",
		StartCmd:    "HOST=0.0.0.0 node .output/server/index.mjs",
		Framework:   "nuxt.js",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	require.Contains(t, result, "RUN NITRO_PRESET=node-server SERVER_PRESET=node-server pnpm build")
	require.Contains(t, result, "if [ -d .zeabur/output ]")
	require.Contains(t, result, "unsupported serverless output")
	require.Contains(t, result, "Use NITRO_PRESET=node-server (or bun) instead of NITRO_PRESET=zeabur / SERVER_PRESET=zeabur.")
}

func TestTemplate_NonNitroOmitsZeaburOutputGuard(t *testing.T) {
	t.Parallel()

	ctx := nodejs.TemplateContext{
		NodeVersion: "18",
		InitCmd:     "RUN npm install -g yarn@latest",
		InstallCmd:  "RUN yarn install",
		BuildCmd:    "yarn build",
		StartCmd:    "yarn start",
		Framework:   "next.js",
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	require.NotContains(t, result, ".zeabur/output")
}
