// Package dotnet is the planner of Dotnet projects.
package dotnet

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// TemplateContext is the context for the Dotnet Dockerfile template.
type TemplateContext struct {
	DotnetVer    string
	Out          string
	Static       bool
	SubmoduleDir string
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

// GenerateDockerfile generates the Dockerfile for Dotnet projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	context := TemplateContext{
		DotnetVer:    meta["sdk"],
		Out:          strings.TrimSuffix(meta["entryPoint"], ".csproj"),
		SubmoduleDir: meta["submoduleDir"],
	}

	if framework := meta["framework"]; framework == "blazorwasm" {
		context.Static = true
	} else {
		context.Static = false
	}

	return context.Execute()
}

type pack struct {
	*identify
}

// NewPacker returns a new Dotnet packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
