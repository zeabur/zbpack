package nodejs

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/zeabur/zbpack/pkg/types"
)

type TemplateContext struct {
	NodeVersion string

	InstallCmd string
	BuildCmd   string
	StartCmd   string

	OutputDir string
	SPA       bool
}

//go:embed templates
var tmplFs embed.FS

var tmpl = template.Must(
	template.New("template.Dockerfile").
		ParseFS(tmplFs, "templates/*"),
)

func (c TemplateContext) Execute() (string, error) {
	writer := new(bytes.Buffer)
	err := tmpl.Execute(writer, c)

	return writer.String(), err
}

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	context := TemplateContext{
		NodeVersion: meta["nodeVersion"],
		InstallCmd:  meta["installCmd"],
		BuildCmd:    meta["buildCmd"],
		StartCmd:    meta["startCmd"],
		OutputDir:   "",
		SPA:         true,
	}

	framework := meta["framework"]
	mpaFrameworks := []types.NodeProjectFramework{
		types.NodeProjectFrameworkHexo,
		types.NodeProjectFrameworkVitepress,
		types.NodeProjectFrameworkAstroStatic,
	}

	if outputDir, ok := meta["outputDir"]; ok {
		context.OutputDir = outputDir

		for _, f := range mpaFrameworks {
			if framework == string(f) {
				context.SPA = false
				break
			}
		}
	}

	return context.Execute()
}
