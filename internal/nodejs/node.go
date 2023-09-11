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

	OutputDir string
	SPA       bool
	Bun       bool
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

func isMpaFramework(framework string) bool {
	mpaFrameworks := []types.NodeProjectFramework{
		types.NodeProjectFrameworkHexo,
		types.NodeProjectFrameworkVitepress,
		types.NodeProjectFrameworkAstroStatic,
		types.NodeProjectFrameworkSolidStartStatic,
	}

	for _, f := range mpaFrameworks {
		if framework == string(f) {
			return true
		}
	}

	return false
}

// isNotMpaFramework is `!isMpaFramework()`, but it's easier to read
func isNotMpaFramework(framework string) bool {
	return !isMpaFramework(framework)
}

func getContextBasedOnMeta(meta types.PlanMeta) TemplateContext {
	context := TemplateContext{
		NodeVersion: meta["nodeVersion"],
		InstallCmd:  meta["installCmd"],
		BuildCmd:    meta["buildCmd"],
		StartCmd:    meta["startCmd"],
		OutputDir:   "",
		SPA:         true,

		// The flag specific to planner/bun.
		Bun: meta["bun"] == "true",
	}

	if outputDir, ok := meta["outputDir"]; ok {
		context.OutputDir = outputDir
		context.SPA = isNotMpaFramework(meta["framework"])
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
