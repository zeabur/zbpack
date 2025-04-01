// Package elixir is the planner of elixir projects.
package elixir

import (
	"bytes"
	"embed"
	"strconv"
	"text/template"

	"github.com/salamer/zbpack/pkg/packer"
	"github.com/salamer/zbpack/pkg/types"
)

// TemplateContext is the context for the Elixir Dockerfile template.
type TemplateContext struct {
	ElixirVer     string
	ElixirPhoenix bool
	ElixirEcto    bool
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

// GenerateDockerfile generates the Dockerfile for Elixir projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	context := TemplateContext{
		ElixirVer: meta["ver"],
	}

	if ElixirFramework := meta["framework"]; ElixirFramework == "phoenix" {
		context.ElixirPhoenix = true
	} else {
		context.ElixirPhoenix = false
	}

	if usesEcto, _ := strconv.ParseBool(meta["ecto"]); usesEcto {
		context.ElixirEcto = true
	} else {
		context.ElixirEcto = false
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
