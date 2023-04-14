package dockerfile

import (
	"github.com/moznion/go-optional"
	"github.com/zeabur/zbpack/pkg/types"
	"os"
	"path"
	"strings"
)

type dockerfilePlanContext struct {
	ExposePort optional.Option[string]
}

type GetMetaOptions struct {
	AbsPath string
}

func GetExposePort(ctx *dockerfilePlanContext, absPath string) string {
	pm := &ctx.ExposePort
	if port, err := pm.Take(); err == nil {
		return port
	}

	filenames := []string{"Dockerfile", "dockerfile"}
	for _, filename := range filenames {
		if _, err := os.Stat(path.Join(absPath, filename)); err == nil {

			content, err := os.ReadFile(path.Join(absPath, filename))
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
	exposePort := GetExposePort(ctx, opt.AbsPath)
	meta := types.PlanMeta{
		"expose": exposePort,
	}
	return meta
}
