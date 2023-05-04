package dockerfile

import (
	"github.com/moznion/go-optional"
	"github.com/zeabur/zbpack/internal/source"
	"github.com/zeabur/zbpack/pkg/types"
	"strings"
)

type dockerfilePlanContext struct {
	src        *source.Source
	ExposePort optional.Option[string]
}

type GetMetaOptions struct {
	Src *source.Source
}

func GetExposePort(ctx *dockerfilePlanContext) string {
	pm := &ctx.ExposePort
	src := *ctx.src
	if port, err := pm.Take(); err == nil {
		return port
	}

	filenames := []string{"Dockerfile", "dockerfile"}
	for _, filename := range filenames {
		if src.HasFile(filename) {
			content, err := src.ReadFile(filename)
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
