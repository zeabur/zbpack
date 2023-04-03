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
		if isSinglePageApp {
			return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + outputDir + ` /static
RUN echo "server { listen 8080; root /static; location / {try_files \$uri /index.html; }}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
		}
		return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + outputDir + ` /usr/share/nginx/html
RUN echo "server { listen 8080; root /usr/share/nginx/html; }"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
	}

	return `FROM node:` + nodeVersion + ` 
ENV PORT=8080
WORKDIR /src
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `
EXPOSE 8080
CMD ` + startCmd, nil
}
