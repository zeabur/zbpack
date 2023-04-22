package rust

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	_ "embed"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed template.Dockerfile
var templateDockerfile string

type GetMetaOptions struct {
	AbsPath string

	// In Rust, the submodule name is the binary name.
	SubmoduleName string
}

func needOpenssl(fs afero.Fs) bool {
	for _, file := range []string{"Cargo.toml", "Cargo.lock"} {
		file, err := fs.Open(file)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Println(err)
			}
			continue
		}
		defer file.Close()

		buf := bufio.NewScanner(file)
		for buf.Scan() {
			if strings.Contains(buf.Text(), "openssl") {
				return true
			}
		}
	}
	return false
}

func GetMeta(options GetMetaOptions) types.PlanMeta {
	fs := afero.NewBasePathFs(afero.NewOsFs(), options.AbsPath)

	var opensslFlag string
	if needOpenssl(fs) {
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
