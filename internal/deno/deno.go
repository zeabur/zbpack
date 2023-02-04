package deno

import (
	"github.com/zeabur/zbpack/pkg/types"
)


func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	//TODO: deno task start: deno.json start task 
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
			dockerfile += `
			CMD ["run", "--allow-net", "` + entry + `"]`
	}
	return dockerfile, nil
}	