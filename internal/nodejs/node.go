package nodejs

import (
	"bytes"
	"html/template"

	"github.com/zeabur/zbpack/pkg/types"
)

type TemplateContext struct {
	NodeVersion string

	InstallCmd string
	BuildCmd   string
	StartCmd   string

	OutputDir string
	SSR       bool
}

var tmpl = template.Must(
	template.New("template.Dockerfile").
		ParseFiles("./template.Dockerfile", "./templates/nginx-runtime.Dockerfile"),
)

func (c TemplateContext) Execute() (string, error) {
	writer := new(bytes.Buffer)
	err := tmpl.Execute(writer, c)

	return writer.String(), err
}

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

	copyLockFile := ""
	switch meta["packageManager"] {
	case string(types.NodePackageManagerNpm):
		copyLockFile = "COPY package-lock.json ."
	case string(types.NodePackageManagerYarn):
		copyLockFile = "COPY yarn.lock ."
	case string(types.NodePackageManagerPnpm):
		copyLockFile = "COPY pnpm-lock.yaml ."
	}

	if outputDir, ok := meta["outputDir"]; ok {

		tryFiles := `try_files \$uri \$uri.html \$uri/index.html /404.html =404;`
		if isSinglePageApp {
			tryFiles = `try_files \$uri /index.html;`
		}

		return `FROM docker.io/library/node:` + nodeVersion + ` as build
WORKDIR /src
COPY package.json .
` + copyLockFile + `
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `

FROM docker.io/library/nginx:alpine as runtime
COPY --from=build /src/` + outputDir + ` /usr/share/nginx/html/static
RUN echo "server { listen 8080; root /usr/share/nginx/html/static; location / {` + tryFiles + `}}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
	}

	return `FROM docker.io/library/node:` + nodeVersion + `
ENV PORT=8080
WORKDIR /src
COPY package.json .
` + copyLockFile + `
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `
EXPOSE 8080
CMD ` + startCmd, nil
}
