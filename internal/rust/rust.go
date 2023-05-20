// Package rust is the build planner for Rust projects.
package rust

import (
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	_ "embed"

	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed template.Dockerfile
var templateDockerfile string

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src afero.Fs

	// In Rust, the submodule name is the binary name.
	SubmoduleName string
}

func needOpenssl(source afero.Fs) bool {
	for _, file := range []string{"Cargo.toml", "Cargo.lock"} {
		file, err := afero.ReadFile(source, file)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Println(err)
			}
			continue
		}

		if strings.Contains(string(file), "openssl") {
			return true
		}
	}
	return false
}

// GetMeta gets the metadata of the Rust project.
func GetMeta(options GetMetaOptions) types.PlanMeta {
	var opensslFlag string
	if needOpenssl(options.Src) {
		opensslFlag = "yes"
	} else {
		opensslFlag = "no"
	}

	return types.PlanMeta{
		"BinName":     options.SubmoduleName,
		"NeedOpenssl": opensslFlag,
	}
}

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
