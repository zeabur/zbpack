package nodejs

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DeterminePackageManager(ctx context.Context, absPath string) NodePackageManager {

	if packageManager, ok := ctx.Value("packageManager").(NodePackageManager); ok {
		return packageManager
	}

	if _, err := os.Stat(path.Join(absPath, "yarn.lock")); err == nil {
		context.WithValue(ctx, "packageManager", NodePackageManagerYarn)
		return NodePackageManagerYarn
	}

	if _, err := os.Stat(path.Join(absPath, "pnpm-lock.yaml")); err == nil {
		context.WithValue(ctx, "packageManager", NodePackageManagerPnpm)
		return NodePackageManagerPnpm
	}

	if _, err := os.Stat(path.Join(absPath, "package-lock.json")); err == nil {
		context.WithValue(ctx, "packageManager", NodePackageManagerNpm)
		return NodePackageManagerNpm
	}

	context.WithValue(ctx, "packageManager", NodePackageManagerYarn)
	return NodePackageManagerYarn
}

func DetermineProjectFramework(ctx context.Context, absPath string) NodeProjectFramework {

	if framework, ok := ctx.Value("framework").(NodeProjectFramework); ok {
		return framework
	}

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNone)
		return NodeProjectFrameworkNone
	}

	packageJson := struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNone)
		return NodeProjectFrameworkNone
	}

	if _, isAstro := packageJson.Dependencies["astro"]; isAstro {
		if _, isAstroSSR := packageJson.Dependencies["@astrojs/node"]; isAstroSSR {
			context.WithValue(ctx, "framework", NodeProjectFrameworkAstroSSR)
			return NodeProjectFrameworkAstroSSR
		}
		context.WithValue(ctx, "framework", NodeProjectFrameworkAstroStatic)
		return NodeProjectFrameworkAstroStatic
	}

	if _, isSvelte := packageJson.DevDependencies["svelte"]; isSvelte {
		context.WithValue(ctx, "framework", NodeProjectFrameworkSvelte)
		return NodeProjectFrameworkSvelte
	}

	if _, isHexo := packageJson.Dependencies["hexo"]; isHexo {
		context.WithValue(ctx, "framework", NodeProjectFrameworkHexo)
		return NodeProjectFrameworkHexo
	}

	if _, isQwik := packageJson.DevDependencies["@builder.io/qwik"]; isQwik {
		context.WithValue(ctx, "framework", NodeProjectFrameworkQwik)
		return NodeProjectFrameworkQwik
	}

	if _, isVitepress := packageJson.DevDependencies["vitepress"]; isVitepress {
		context.WithValue(ctx, "framework", NodeProjectFrameworkVitepress)
		return NodeProjectFrameworkVitepress
	}

	if _, isVite := packageJson.DevDependencies["vite"]; isVite {
		context.WithValue(ctx, "framework", NodeProjectFrameworkVite)
		return NodeProjectFrameworkVite
	}

	if _, isUmi := packageJson.Dependencies["umi"]; isUmi {
		context.WithValue(ctx, "framework", NodeProjectFrameworkUmi)
		return NodeProjectFrameworkUmi
	}

	if _, isNextJs := packageJson.Dependencies["next"]; isNextJs {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNextJs)
		return NodeProjectFrameworkNextJs
	}

	if _, isNestJs := packageJson.Dependencies["@nestjs/core"]; isNestJs {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNestJs)
		return NodeProjectFrameworkNestJs
	}

	if _, isRemix := packageJson.Dependencies["@remix-run/react"]; isRemix {
		context.WithValue(ctx, "framework", NodeProjectFrameworkRemix)
		return NodeProjectFrameworkRemix
	}

	if _, isCreateReactApp := packageJson.Dependencies["react-scripts"]; isCreateReactApp {
		context.WithValue(ctx, "framework", NodeProjectFrameworkCreateReactApp)
		return NodeProjectFrameworkCreateReactApp
	}

	if _, isNuxtJs := packageJson.Dependencies["nuxt"]; isNuxtJs {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNuxtJs)
		return NodeProjectFrameworkNuxtJs
	}

	if _, isNuxtJs := packageJson.DevDependencies["nuxt"]; isNuxtJs {
		context.WithValue(ctx, "framework", NodeProjectFrameworkNuxtJs)
		return NodeProjectFrameworkNuxtJs
	}

	if _, isVueCliApp := packageJson.DevDependencies["@vue/cli-service"]; isVueCliApp {
		context.WithValue(ctx, "framework", NodeProjectFrameworkVueCli)
		return NodeProjectFrameworkVueCli
	}

	context.WithValue(ctx, "framework", NodeProjectFrameworkNone)
	return NodeProjectFrameworkNone
}

func DetermineNeedPuppeteer(ctx context.Context, absPath string) bool {

	if needPuppeteer, ok := ctx.Value("needPuppeteer").(bool); ok {
		return needPuppeteer
	}

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		context.WithValue(ctx, "needPuppeteer", false)
		return false
	}

	packageJson := struct {
		Dependencies map[string]string `json:"dependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		context.WithValue(ctx, "needPuppeteer", false)
		return false
	}

	if _, hasPuppeteer := packageJson.Dependencies["puppeteer"]; hasPuppeteer {
		context.WithValue(ctx, "needPuppeteer", true)
		return true
	}

	context.WithValue(ctx, "needPuppeteer", false)
	return false
}

func GetBuildScript(ctx context.Context, absPath string) string {

	if buildScript, ok := ctx.Value("buildScript").(string); ok {
		return buildScript
	}

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		context.WithValue(ctx, "buildScript", "")
		return ""
	}

	packageJson := struct {
		Scripts map[string]string `json:"scripts"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		context.WithValue(ctx, "buildScript", "")
		return ""
	}

	if _, ok := packageJson.Scripts["build"]; ok {
		context.WithValue(ctx, "buildScript", "build")
		return "build"
	}

	for key := range packageJson.Scripts {
		if strings.Contains(key, "build") {
			context.WithValue(ctx, "buildScript", key)
			return key
		}
	}

	context.WithValue(ctx, "buildScript", "")
	return ""
}

func GetStartScript(ctx context.Context, absPath string) string {

	if startScript, ok := ctx.Value("startScript").(string); ok {
		return startScript
	}

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		context.WithValue(ctx, "startScript", "")
		return ""
	}

	packageJson := struct {
		Scripts         map[string]string `json:"scripts"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		context.WithValue(ctx, "startScript", "")
		return ""
	}

	if _, ok := packageJson.DevDependencies["@builder.io/qwik"]; ok {
		if _, ok := packageJson.Scripts["deploy"]; ok {
			context.WithValue(ctx, "startScript", "deploy")
			return "deploy"
		}
	}

	if _, ok := packageJson.Scripts["start"]; ok {
		context.WithValue(ctx, "startScript", "start")
		return "start"
	}

	context.WithValue(ctx, "startScript", "")
	return ""
}

func GetNodeVersion(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return "16"
	}

	if packageJson.Engines.Node == "" {
		return "16"
	}

	// for example, ">=16.0.0 <17.0.0"
	versionRange := packageJson.Engines.Node

	isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	if isVersion {
		return versionRange
	}

	// from given version range, we want to extract the minimum version
	// for example, "16"
	ranges := strings.Split(versionRange, " ")
	minVer := -1
	equalMin := false
	maxVer := -1
	equalMax := false
	for _, r := range ranges {
		if strings.HasPrefix(r, ">=") {
			minVerString := strings.TrimPrefix(r, ">=")
			minVerMajor := strings.Split(minVerString, ".")[0]
			minVer, _ = strconv.Atoi(minVerMajor)
			equalMin = true
		} else if strings.HasPrefix(r, ">") {
			minVerString := strings.TrimPrefix(r, ">")
			minVerMajor := strings.Split(minVerString, ".")[0]
			minVer, _ = strconv.Atoi(minVerMajor)
			equalMin = false
		} else if strings.HasPrefix(r, "<=") {
			maxVerString := strings.TrimPrefix(r, "<=")
			maxVerMajor := strings.Split(maxVerString, ".")[0]
			maxVer, _ = strconv.Atoi(maxVerMajor)
			equalMax = true
		} else if strings.HasPrefix(r, "<") {
			maxVerString := strings.TrimPrefix(r, "<")
			maxVerMajor := strings.Split(maxVerString, ".")[0]
			maxVer, _ = strconv.Atoi(maxVerMajor)
			equalMax = false
		}
	}

	if minVer == -1 && maxVer == -1 {
		return "16"
	}

	if minVer == -1 {
		if equalMax {
			if maxVer != 14 && maxVer != 16 && maxVer != 18 {
				return "16"
			}
			return strconv.Itoa(maxVer)
		} else {
			return strconv.Itoa(maxVer - 1)
		}
	}

	if maxVer == -1 {
		if equalMin {
			if minVer != 14 && minVer != 16 && minVer != 18 {
				return "16"
			}
			return strconv.Itoa(minVer)
		} else {
			return strconv.Itoa(minVer + 1)
		}
	}

	return "16"
}

func GetEntry(ctx context.Context, absPath string) string {
	if entry, ok := ctx.Value("entry").(string); ok {
		return entry
	}

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		context.WithValue(ctx, "entry", "")
		return ""
	}

	packageJson := struct {
		Main string `json:"main"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		context.WithValue(ctx, "entry", "")
		return ""
	}

	context.WithValue(ctx, "entry", packageJson.Main)
	return packageJson.Main
}

func GetInstallCmd(ctx context.Context, absPath string) string {

	if installCmd, ok := ctx.Value("installCmd").(string); ok {
		return installCmd
	}

	pkgManager := DeterminePackageManager(ctx, absPath)
	installCmd := "yarn"
	switch pkgManager {
	case NodePackageManagerNpm:
		installCmd = "npm install"
	case NodePackageManagerYarn:
		installCmd = "yarn install"
	case NodePackageManagerPnpm:
		installCmd = "npm install -g pnpm && pnpm install"
	}

	context.WithValue(ctx, "installCmd", installCmd)
	return installCmd
}

func GetBuildCmd(ctx context.Context, absPath string) string {

	if buildCmd, ok := ctx.Value("buildCmd").(string); ok {
		return buildCmd
	}

	buildScript := GetBuildScript(ctx, absPath)
	pkgManager := DeterminePackageManager(ctx, absPath)

	buildCmd := "yarn " + buildScript
	switch pkgManager {
	case NodePackageManagerYarn:
		buildCmd = "yarn " + buildScript
	case NodePackageManagerPnpm:
		buildCmd = "pnpm run " + buildScript
	case NodePackageManagerNpm:
		buildCmd = "npm run " + buildScript
	}

	if buildScript == "" {
		buildCmd = ""
	}

	needPuppeteer := DetermineNeedPuppeteer(ctx, absPath)
	if needPuppeteer {
		buildCmd = `apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1 && groupadd -r puppeteer && useradd -r -g puppeteer -G audio,video puppeteer && chown -R puppeteer:puppeteer /src && mkdir /home/puppeteer && chown -R puppeteer:puppeteer /home/puppeteer && USER puppeteer && ` + buildCmd
	}

	context.WithValue(ctx, "buildCmd", buildCmd)
	return buildCmd
}

func GetStartCmd(ctx context.Context, absPath string) string {

	if startCmd, ok := ctx.Value("startCmd").(string); ok {
		return startCmd
	}

	startScript := GetStartScript(ctx, absPath)
	pkgManager := DeterminePackageManager(ctx, absPath)
	entry := GetEntry(ctx, absPath)
	framework := DetermineProjectFramework(ctx, absPath)

	startCmd := "yarn " + startScript
	switch pkgManager {
	case NodePackageManagerYarn:
		startCmd = "yarn " + startScript
	case NodePackageManagerPnpm:
		startCmd = "pnpm " + startScript
	case NodePackageManagerNpm:
		startCmd = "npm run " + startScript
	}

	if startScript == "" {
		if entry != "" {
			startCmd = "node " + entry
		} else if framework == NodeProjectFrameworkNuxtJs {
			startCmd = "node .output/server/index.mjs"
		} else {
			startCmd = "node index.js"
		}
	}

	needPuppeteer := DetermineNeedPuppeteer(ctx, absPath)
	if needPuppeteer {
		startCmd = "node node_modules/puppeteer/install.js && " + startCmd
	}

	context.WithValue(ctx, "startCmd", startCmd)
	return startCmd
}

func GetMeta(absPath string) PlanMeta {
	ctx := context.TODO()
	framework := DetermineProjectFramework(ctx, absPath)
	nodeVersion := GetNodeVersion(absPath)

	installCmd := GetInstallCmd(ctx, absPath)
	buildCmd := GetBuildCmd(ctx, absPath)
	startCmd := GetStartCmd(ctx, absPath)

	return PlanMeta{
		"framework":   string(framework),
		"nodeVersion": nodeVersion,
		"installCmd":  installCmd,
		"buildCmd":    buildCmd,
		"startCmd":    startCmd,
	}
}
