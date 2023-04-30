package python

import (
	"github.com/zeabur/zbpack/pkg/types"
)

// IsMysqlNeeded checks if the project has a dependency on `mysqlclient`,
// and it will return true if it does.

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	installCmd := meta["install"]
	startCmd := meta["start"]
	needMySQL := meta["needMySQL"]

	dockerfile := "FROM docker.io/library/python:3.8.2-slim-buster\n"

	if needMySQL == "true" {
		dockerfile += `RUN apt update \
	&& apt install -y libmariadb-dev build-essential \
	&& rm -rf /var/lib/apt/lists/*`
	}

	dockerfile += `WORKDIR /app
RUN apt-get update && apt-get install gcc -y
COPY . .
RUN ` + installCmd + `
EXPOSE 8080
CMD ` + startCmd

	return dockerfile, nil
}
