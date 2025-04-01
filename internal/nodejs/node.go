// Package nodejs generates the Dockerfile for Node.js projects.
package nodejs

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"github.com/salamer/zbpack/pkg/packer"
	"github.com/salamer/zbpack/pkg/types"
)

// TemplateContext is the context for the Node.js Dockerfile template.
type TemplateContext struct {
	NodeVersion string

	AppDir string

	InitCmd    string
	InstallCmd string
	BuildCmd   string
	StartCmd   string

	Framework string
	OutputDir string
}

//go:embed templates
var tmplFs embed.FS

var tmpl = template.Must(
	template.New("template.Dockerfile").
		Funcs(template.FuncMap{
			"prefixed": strings.HasPrefix,
			"isNitro":  types.IsNitroBasedFramework,
		}).
		ParseFS(tmplFs, "templates/*"),
)

// Execute executes the template.
func (c TemplateContext) Execute() (string, error) {
	writer := new(bytes.Buffer)
	err := tmpl.Execute(writer, c)

	return writer.String(), err
}

func getContextBasedOnMeta(meta types.PlanMeta) TemplateContext {
	context := TemplateContext{
		NodeVersion: meta["nodeVersion"],
		AppDir:      meta["appDir"],
		InitCmd:     meta["initCmd"],
		InstallCmd:  meta["installCmd"],
		BuildCmd:    meta["buildCmd"],
		StartCmd:    meta["startCmd"],
		Framework:   meta["framework"],
		OutputDir:   meta["outputDir"],
	}

	return context
}

// GenerateDockerfile generates the Dockerfile for Node.js projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return getContextBasedOnMeta(meta).Execute()
}

type pack struct {
	*identify
}

// NewPacker returns a new Node.js packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
