// Package nix is the packer for Nix projects.
package nix

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed Dockerfile.tmpl
var tmplRaw string

// GenerateDockerfile generates the Dockerfile for Nix projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	type TemplateContext struct {
		Package string
	}

	tmpl, err := template.New("Dockerfile").Parse(tmplRaw)
	if err != nil {
		return "", err
	}

	out := new(bytes.Buffer)

	err = tmpl.Execute(out, TemplateContext{
		Package: meta["package"],
	})
	if err != nil {
		return "", err
	}

	return out.String(), nil
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
