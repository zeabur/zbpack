package deno

import (
	"github.com/zeabur/zbpack/pkg/types"
)


func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	framework := meta["framework"]
	entry := meta["entry"]
	startCmd := meta["startCommand"]

	dockerfile := `FROM denoland/deno
	WORKDIR /app
	COPY . .
	EXPOSE 8080
	RUN deno cache ` + entry

	switch(framework){
		case string(types.DenoFrameworkFresh):
			dockerfile += `
			CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-run", "` + entry + `"]`
		case string(types.DenoFrameworkNone):
			if startCmd == "" {
				dockerfile += `
				CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-run", "` + entry + `"]`
			} else {
				dockerfile += `
				CMD ["deno", "task", "start"]` 
			}
	}
	return dockerfile, nil
}