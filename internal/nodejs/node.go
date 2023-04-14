package nodejs

import (
	"github.com/zeabur/zbpack/pkg/types"
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

		return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY . .
RUN ` + installCmd + `
RUN ` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + outputDir + ` /usr/share/nginx/html
RUN echo "server { listen 8080; root /usr/share/nginx/html; location / {` + tryFiles + `}}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
	}

	return `FROM node:` + nodeVersion + ` 
ENV PORT=8080
WORKDIR /src
COPY . .
RUN ` + installCmd + `
RUN ` + buildCmd + `
EXPOSE 8080
CMD ` + startCmd, nil
}
