package nodejs

import "github.com/zeabur/zbpack/pkg/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {

	pkgManager := meta["packageManager"]

	installCmd := "RUN npm install"
	switch pkgManager {
	case string(types.NodePackageManagerYarn):
		installCmd = "RUN yarn install"
	case string(types.NodePackageManagerPnpm):
		installCmd = `
RUN npm install -g pnpm
RUN pnpm install
`
	}

	buildCmd := "RUN npm run " + meta["buildCommand"]
	switch pkgManager {
	case string(types.NodePackageManagerYarn):
		buildCmd = "RUN yarn " + meta["buildCommand"]
	case string(types.NodePackageManagerPnpm):
		buildCmd = "RUN pnpm run " + meta["buildCommand"]
	}
	if meta["buildCommand"] == "" {
		buildCmd = ""
	}

	startCmd := "CMD npm run " + meta["startCommand"]
	switch pkgManager {
	case string(types.NodePackageManagerYarn):
		startCmd = "CMD yarn " + meta["startCommand"]
	case string(types.NodePackageManagerPnpm):
		startCmd = "CMD pnpm " + meta["startCommand"]
	}
	if meta["startCommand"] == "" {
		if meta["mainFile"] != "" {
			startCmd = "CMD node " + meta["mainFile"]
		} else {
			startCmd = "CMD node index.js"
		}
	}

	framework := meta["framework"]

	nodeVersion := meta["nodeVersion"]

	// TODO: get isStaticOutput from meta
	isStaticOutput := false

	// TODO: get staticOutputDir from meta
	staticOutputDir := ""

	staticFrameworks := []types.NodeProjectFramework{
		types.NodeProjectFrameworkVite,
		types.NodeProjectFrameworkUmi,
		types.NodeProjectFrameworkCreateReactApp,
		types.NodeProjectFrameworkVueCli,
	}

	defaultStaticOutputDirs := map[types.NodeProjectFramework]string{
		types.NodeProjectFrameworkVite:           "dist",
		types.NodeProjectFrameworkUmi:            "dist",
		types.NodeProjectFrameworkVueCli:         "dist",
		types.NodeProjectFrameworkCreateReactApp: "build",
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
		return `FROM node:` + nodeVersion + ` as build
WORKDIR /src
COPY . .
` + installCmd + `
` + buildCmd + `

FROM nginx:alpine
COPY --from=build /src/` + staticOutputDir + ` /static
RUN echo "server { listen 8080; root /static; location / {try_files \$uri /index.html; }}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
`, nil
	}

	puppeteerCmd := ""
	needPuppeteer := meta["needPuppeteer"] == "true"
	if needPuppeteer {
		puppeteerCmd = `
RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1
RUN groupadd -r puppeteer 
RUN useradd -r -g puppeteer -G audio,video puppeteer
RUN chown -R puppeteer:puppeteer /src
RUN mkdir /home/puppeteer && chown -R puppeteer:puppeteer /home/puppeteer
USER puppeteer
`

		startCmd = "CMD " + "node node_modules/puppeteer/install.js && " + startCmd[4:]
	}

	return `FROM node:` + nodeVersion + ` 
ENV PORT=8080
WORKDIR /src
COPY . .
` + installCmd + `
` + buildCmd + `
` + puppeteerCmd + `
EXPOSE 8080
` + startCmd, nil
}
