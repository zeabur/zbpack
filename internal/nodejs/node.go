package nodejs

import (
	"github.com/zeabur/zbpack/pkg/types"
	"strings"
)

func GenerateDockerfile(meta types.PlanMeta) (string, error) {

	framework := meta["framework"]
	nodeVersion := meta["nodeVersion"]
	installCmd := meta["installCmd"]
	buildCmd := meta["buildCmd"]
	startCmd := meta["startCmd"]

	// TODO: get isSinglePageApp from meta
	isSinglePageApp := true
	if framework == string(types.NodeProjectFrameworkHexo) || framework == string(types.NodeProjectFrameworkVitepress) || framework == string(types.NodeProjectFrameworkAstroStatic) {
		isSinglePageApp = false
	}

	if outputDir, ok := meta["outputDir"]; ok {

		tryFiles := `try_files \$uri \$uri.html \$uri/index.html /404.html =404;`
		if isSinglePageApp {
			tryFiles = `try_files \$uri /index.html;`
		}

		return `FROM docker.io/library/node:` + nodeVersion + ` as build
WORKDIR /src
COPY . .
RUN ` + installCmd + `
RUN ` + buildCmd + `

FROM docker.io/library/nginx:alpine as runtime 
COPY --from=build /src/` + outputDir + ` /usr/share/nginx/html/static
RUN echo "server { listen 8080; root /usr/share/nginx/html/static; location / {` + tryFiles + `}}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
	}

	lockfile := "package-lock.json"
	if strings.Contains(installCmd, "yarn") {
		lockfile = "yarn.lock"
	}
	if strings.Contains(installCmd, "pnpm") {
		lockfile = "pnpm-lock.yaml"
	}

	return `FROM docker.io/library/node:` + nodeVersion + ` 
ENV PORT=8080
WORKDIR /src
COPY package.json .
COPY ` + lockfile + ` .
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `
EXPOSE 8080
CMD ` + startCmd, nil
}
