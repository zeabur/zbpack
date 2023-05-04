package rust

import (
	"bytes"
	"github.com/zeabur/zbpack/internal/source"
	"log"
	"os"
	"strings"
	"text/template"

	_ "embed"

	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed template.Dockerfile
var templateDockerfile string

type GetMetaOptions struct {
	Src *source.Source

	// In Rust, the submodule name is the binary name.
	SubmoduleName string
}

func needOpenssl(source *source.Source) bool {
	src := *source
	for _, file := range []string{"Cargo.toml", "Cargo.lock"} {
		file, err := src.ReadFile(file)
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
