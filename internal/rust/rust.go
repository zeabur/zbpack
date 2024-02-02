// Package rust is the build planner for Rust projects.
package rust

import (
	"bytes"
	"text/template"

	_ "embed"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed template.Dockerfile
var templateDockerfile string

// GenerateDockerfile generates the Dockerfile for the Rust project.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	template := template.Must(
		template.New("RustDockerfile").Parse(templateDockerfile),
	)

	var result bytes.Buffer

	if err := template.Execute(&result, meta); err != nil {
		return "", err
	}

	return result.String(), nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Rust packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
