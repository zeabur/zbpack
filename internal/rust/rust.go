// Package rust is the build planner for Rust projects.
package rust

import (
	"bytes"
	"strings"
	"text/template"

	_ "embed"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed template.Dockerfile
var dockerTemplate string

// TemplateContext is the context for the Dockerfile template.
type TemplateContext struct {
	OpenSSL    bool
	Serverless bool
	Entry      string
	AppDir     string
	Assets     []string

	BuildCommand    string
	StartCommand    string
	PreStartCommand string
}

// GenerateDockerfile generates the Dockerfile for the Rust project.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	template := template.Must(
		template.New("RustDockerfile").Parse(dockerTemplate),
	)

	context := TemplateContext{
		OpenSSL:         meta["openssl"] == "true",
		Serverless:      meta["serverless"] == "true",
		Entry:           meta["entry"],
		AppDir:          meta["appDir"],
		Assets:          strings.FieldsFunc(meta["assets"], func(r rune) bool { return r == ':' }),
		BuildCommand:    meta["buildCommand"],
		StartCommand:    meta["startCommand"],
		PreStartCommand: meta["preStartCommand"],
	}

	var result bytes.Buffer

	if err := template.Execute(&result, context); err != nil {
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
