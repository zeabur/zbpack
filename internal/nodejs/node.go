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

	// TODO: get isStaticOutput from meta
	isStaticOutput := false

	// TODO: get staticOutputDir from meta
	staticOutputDir := ""

	// TODO: get isSinglePageApp from meta
	isSinglePageApp := true
	if framework == string(types.NodeProjectFrameworkHexo) || framework == string(types.NodeProjectFrameworkVitepress) || framework == string(types.NodeProjectFrameworkAstroStatic) {
		isSinglePageApp = false
	}

	staticFrameworks := []types.NodeProjectFramework{
		types.NodeProjectFrameworkVite,
		types.NodeProjectFrameworkUmi,
		types.NodeProjectFrameworkCreateReactApp,
		types.NodeProjectFrameworkVueCli,
		types.NodeProjectFrameworkHexo,
		types.NodeProjectFrameworkVitepress,
		types.NodeProjectFrameworkAstroStatic,
	}

	defaultStaticOutputDirs := map[types.NodeProjectFramework]string{
		types.NodeProjectFrameworkVite:           "dist",
		types.NodeProjectFrameworkUmi:            "dist",
		types.NodeProjectFrameworkVueCli:         "dist",
		types.NodeProjectFrameworkCreateReactApp: "build",
		types.NodeProjectFrameworkHexo:           "public",
		types.NodeProjectFrameworkVitepress:      "docs/.vitepress/dist",
		types.NodeProjectFrameworkAstroStatic:    "dist",
	}

	for _, f := range staticFrameworks {
		if framework == string(f) {
			isStaticOutput = true
			if staticOutputDir == "" {
				staticOutputDir = defaultStaticOutputDirs[f]
			}
		}
	}

	if isStaticOutput {
		if isSinglePageApp {
			return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY . .
RUN ` + installCmd + `
RUN ` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + staticOutputDir + ` /static
RUN echo "server { listen 8080; root /static; location / {try_files \$uri /index.html; }}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
		}
		return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY . .
RUN ` + installCmd + `
RUN ` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + staticOutputDir + ` /usr/share/nginx/html
RUN echo "server { listen 8080; root /usr/share/nginx/html; }"> /etc/nginx/conf.d/default.conf
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
