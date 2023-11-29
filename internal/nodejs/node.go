// Package nodejs generates the Dockerfile for Node.js projects.
package nodejs

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// TemplateContext is the context for the Node.js Dockerfile template.
type TemplateContext struct {
	NodeVersion string

	InstallCmd string
	BuildCmd   string
	StartCmd   string

	Framework  string
	Serverless bool
	OutputDir  string

	Bun bool
}

//go:embed templates
var tmplFs embed.FS

var tmpl = template.Must(
	template.New("template.Dockerfile").
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
		InstallCmd:  meta["installCmd"],
		BuildCmd:    meta["buildCmd"],
		StartCmd:    meta["startCmd"],
		Framework:   meta["framework"],
		Serverless:  meta["serverless"] == "true",
		OutputDir:   meta["outputDir"],

		// The flag specific to planner/bun.
		Bun: meta["bun"] == "true",
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
