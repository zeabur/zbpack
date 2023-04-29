package nodejs_test

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

var tmpl = template.Must(
	template.New("template.Dockerfile").
		ParseFiles("./template.Dockerfile", "./templates/nginx-runtime.Dockerfile"),
)

type TemplateContext struct {
	NodeVersion string

	InstallCmd string
	BuildCmd   string
	StartCmd   string

	OutputDir string
	SSR       bool
}

func (c TemplateContext) Execute() (string, error) {
	writer := new(bytes.Buffer)
	err := tmpl.Execute(writer, c)

	return writer.String(), err
}

func TestTemplate_NBuildCmd_NOutputDir(t *testing.T) {
	ctx := TemplateContext{
		NodeVersion: "18",

		InstallCmd: "yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",

		OutputDir: "",
		SSR:       false,
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_NBuildCmd_OutputDir_NSSR(t *testing.T) {
	ctx := TemplateContext{
		NodeVersion: "18",

		InstallCmd: "yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",

		OutputDir: "dist",
		SSR:       false,
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_NBuildCmd_OutputDir_SSR(t *testing.T) {
	ctx := TemplateContext{
		NodeVersion: "18",

		InstallCmd: "yarn install",
		BuildCmd:   "",
		StartCmd:   "yarn start",

		OutputDir: "dist",
		SSR:       true,
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_NOutputDir(t *testing.T) {
	ctx := TemplateContext{
		NodeVersion: "18",

		InstallCmd: "yarn install",
		BuildCmd:   "yarn build",
		StartCmd:   "yarn start",

		OutputDir: "",
		SSR:       false,
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}

func TestTemplate_BuildCmd_OutputDir(t *testing.T) {
	ctx := TemplateContext{
		NodeVersion: "18",

		InstallCmd: "yarn install",
		BuildCmd:   "yarn build",
		StartCmd:   "yarn start",

		OutputDir: "dist",
		SSR:       true,
	}

	result, err := ctx.Execute()
	assert.NoError(t, err)
	snaps.MatchSnapshot(t, result)
}
