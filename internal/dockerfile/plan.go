package dockerfile

import (
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

type dockerfilePlanContext struct {
	src        afero.Fs
	ExposePort optional.Option[string]
}

type GetMetaOptions struct {
	Src afero.Fs
}

func GetExposePort(ctx *dockerfilePlanContext) string {
	pm := &ctx.ExposePort
	src := ctx.src
	if port, err := pm.Take(); err == nil {
		return port
	}

	filenames := []string{"Dockerfile", "dockerfile"}
	for _, filename := range filenames {
		if utils.HasFile(src, filename) {
			content, err := afero.ReadFile(src, filename)
			if err != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.ToUpper(line), "EXPOSE") {
					v := strings.Split(line, " ")[1]
					*pm = optional.Some(v)
					return pm.Unwrap()
				}
			}

		}
	}

	*pm = optional.Some("8080")
	return pm.Unwrap()
}

func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := new(dockerfilePlanContext)
	ctx.src = opt.Src
	exposePort := GetExposePort(ctx)
	meta := types.PlanMeta{
		"expose": exposePort,
	}
	return meta
}
