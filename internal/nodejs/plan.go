package nodejs

import (
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	. "github.com/zeabur/zbpack/pkg/types"
)

type nodePlanContext struct {
	PackageJson PackageJson

	PackageManager  optional.Option[NodePackageManager]
	Framework       optional.Option[NodeProjectFramework]
	NeedPuppeteer   optional.Option[bool]
	BuildScript     optional.Option[string]
	StartScript     optional.Option[string]
	Entry           optional.Option[string]
	InstallCmd      optional.Option[string]
	BuildCmd        optional.Option[string]
	StartCmd        optional.Option[string]
	StaticOutputDir optional.Option[string]
	// ...
}

func DeterminePackageManager(ctx *nodePlanContext, absPath string) NodePackageManager {
	pm := &ctx.PackageManager

	if packageManager, err := pm.Take(); err == nil {
		return packageManager
	}

	if _, err := os.Stat(path.Join(absPath, "yarn.lock")); err == nil {
		*pm = optional.Some(NodePackageManagerYarn)
		return pm.Unwrap()
	}

	if _, err := os.Stat(path.Join(absPath, "pnpm-lock.yaml")); err == nil {
		*pm = optional.Some(NodePackageManagerPnpm)
		return pm.Unwrap()
	}

	if _, err := os.Stat(path.Join(absPath, "package-lock.json")); err == nil {
		*pm = optional.Some(NodePackageManagerNpm)
		return pm.Unwrap()
	}

	*pm = optional.Some(NodePackageManagerYarn)
	return pm.Unwrap()
}

func DetermineProjectFramework(ctx *nodePlanContext) NodeProjectFramework {
	fw := &ctx.Framework
	packageJson := ctx.PackageJson

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if _, isAstro := packageJson.Dependencies["astro"]; isAstro {
		if _, isAstroSSR := packageJson.Dependencies["@astrojs/node"]; isAstroSSR {
			*fw = optional.Some(NodeProjectFrameworkAstroSSR)
			return fw.Unwrap()
		}

		*fw = optional.Some(NodeProjectFrameworkAstroStatic)
		return fw.Unwrap()
	}

	if _, isSvelte := packageJson.DevDependencies["svelte"]; isSvelte {
		*fw = optional.Some(NodeProjectFrameworkSvelte)
		return fw.Unwrap()
	}

	if _, isHexo := packageJson.Dependencies["hexo"]; isHexo {
		*fw = optional.Some(NodeProjectFrameworkHexo)
		return fw.Unwrap()
	}

	if _, isQwik := packageJson.DevDependencies["@builder.io/qwik"]; isQwik {
		*fw = optional.Some(NodeProjectFrameworkQwik)
		return fw.Unwrap()
	}

	if _, isVitepress := packageJson.DevDependencies["vitepress"]; isVitepress {
		*fw = optional.Some(NodeProjectFrameworkVitepress)
		return fw.Unwrap()
	}

	if _, isVite := packageJson.DevDependencies["vite"]; isVite {
		*fw = optional.Some(NodeProjectFrameworkVite)
		return fw.Unwrap()
	}

	if _, isUmi := packageJson.Dependencies["umi"]; isUmi {
		*fw = optional.Some(NodeProjectFrameworkUmi)
		return fw.Unwrap()
	}

	if _, isNextJs := packageJson.Dependencies["next"]; isNextJs {
		*fw = optional.Some(NodeProjectFrameworkNextJs)
		return fw.Unwrap()
	}

	if _, isNestJs := packageJson.Dependencies["@nestjs/core"]; isNestJs {
		*fw = optional.Some(NodeProjectFrameworkNestJs)
		return fw.Unwrap()
	}

	if _, isRemix := packageJson.Dependencies["@remix-run/react"]; isRemix {
		*fw = optional.Some(NodeProjectFrameworkRemix)
		return fw.Unwrap()
	}

	if _, isCreateReactApp := packageJson.Dependencies["react-scripts"]; isCreateReactApp {
		*fw = optional.Some(NodeProjectFrameworkCreateReactApp)
		return fw.Unwrap()
	}

	if _, isNuxtJs := packageJson.Dependencies["nuxt"]; isNuxtJs {
		*fw = optional.Some(NodeProjectFrameworkNuxtJs)
		return fw.Unwrap()
	}

	if _, isNuxtJs := packageJson.DevDependencies["nuxt"]; isNuxtJs {
		*fw = optional.Some(NodeProjectFrameworkNuxtJs)
		return fw.Unwrap()
	}

	if _, isVueCliApp := packageJson.DevDependencies["@vue/cli-service"]; isVueCliApp {
		*fw = optional.Some(NodeProjectFrameworkVueCli)
		return fw.Unwrap()
	}

	*fw = optional.Some(NodeProjectFrameworkNone)
	return fw.Unwrap()
}

func DetermineNeedPuppeteer(ctx *nodePlanContext) bool {
	pup := &ctx.NeedPuppeteer
	packageJson := ctx.PackageJson

	if needPuppeteer, err := pup.Take(); err == nil {
		return needPuppeteer
	}

	if _, hasPuppeteer := packageJson.Dependencies["puppeteer"]; hasPuppeteer {
		*pup = optional.Some(true)
		return pup.Unwrap()
	}

	*pup = optional.Some(false)
	return pup.Unwrap()
}

func GetBuildScript(ctx *nodePlanContext) string {
	bs := &ctx.BuildScript
	packageJson := ctx.PackageJson

	if buildScript, err := bs.Take(); err == nil {
		return buildScript
	}

	if _, ok := packageJson.Scripts["build"]; ok {
		*bs = optional.Some("build")
		return bs.Unwrap()
	}

	for key := range packageJson.Scripts {
		if strings.Contains(key, "build") {
			*bs = optional.Some(key)
			return bs.Unwrap()
		}
	}

	*bs = optional.Some("")
	return bs.Unwrap()
}

func GetStartScript(ctx *nodePlanContext) string {
	ss := &ctx.StartScript
	packageJson := ctx.PackageJson

	if startScript, err := ss.Take(); err == nil {
		return startScript
	}

	if _, ok := packageJson.DevDependencies["@builder.io/qwik"]; ok {
		if _, ok := packageJson.Scripts["deploy"]; ok {
			*ss = optional.Some("deploy")
			return ss.Unwrap()
		}
	}

	if _, ok := packageJson.Scripts["start"]; ok {
		*ss = optional.Some("start")
		return ss.Unwrap()
	}

	*ss = optional.Some("")
	return ss.Unwrap()
}

func GetNodeVersion(ctx *nodePlanContext) string {
	packageJson := ctx.PackageJson

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

func GetEntry(ctx *nodePlanContext) string {
	packageJson := ctx.PackageJson
	ent := &ctx.Entry

	if entry, err := ent.Take(); err == nil {
		return entry
	}

	*ent = optional.Some(packageJson.Main)
	return ent.Unwrap()
}

func GetInstallCmd(ctx *nodePlanContext, absPath string) string {
	cmd := &ctx.InstallCmd

	if installCmd, err := cmd.Take(); err == nil {
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

	*cmd = optional.Some(installCmd)
	return cmd.Unwrap()
}

func GetBuildCmd(ctx *nodePlanContext, absPath string) string {
	cmd := &ctx.BuildCmd

	if buildCmd, err := cmd.Take(); err == nil {
		return buildCmd
	}

	buildScript := GetBuildScript(ctx)
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

	needPuppeteer := DetermineNeedPuppeteer(ctx)
	if needPuppeteer {
		buildCmd = `apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1 && groupadd -r puppeteer && useradd -r -g puppeteer -G audio,video puppeteer && chown -R puppeteer:puppeteer /src && mkdir /home/puppeteer && chown -R puppeteer:puppeteer /home/puppeteer && USER puppeteer && ` + buildCmd
	}

	*cmd = optional.Some(buildCmd)
	return cmd.Unwrap()
}

func GetStartCmd(ctx *nodePlanContext, absPath string) string {
	cmd := &ctx.StartCmd

	if startCmd, err := cmd.Take(); err == nil {
		return startCmd
	}

	startScript := GetStartScript(ctx)
	pkgManager := DeterminePackageManager(ctx, absPath)
	entry := GetEntry(ctx)
	framework := DetermineProjectFramework(ctx)

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

	needPuppeteer := DetermineNeedPuppeteer(ctx)
	if needPuppeteer {
		startCmd = "node node_modules/puppeteer/install.js && " + startCmd
	}

	*cmd = optional.Some(startCmd)
	return cmd.Unwrap()
}

// GetStaticOutputDir returns the output directory for static projects.
// If empty string is returned, the service is not deployed as static files.
func GetStaticOutputDir(ctx *nodePlanContext) string {
	dir := &ctx.StaticOutputDir

	if outputDir, err := dir.Take(); err == nil {
		return outputDir
	}

	framework := DetermineProjectFramework(ctx)

	defaultStaticOutputDirs := map[NodeProjectFramework]string{
		NodeProjectFrameworkVite:           "dist",
		NodeProjectFrameworkUmi:            "dist",
		NodeProjectFrameworkVueCli:         "dist",
		NodeProjectFrameworkCreateReactApp: "build",
		NodeProjectFrameworkHexo:           "public",
		NodeProjectFrameworkVitepress:      "docs/.vitepress/dist",
		NodeProjectFrameworkAstroStatic:    "dist",
	}

	if outputDir, ok := defaultStaticOutputDirs[framework]; ok {
		*dir = optional.Some(outputDir)
		return dir.Unwrap()
	}

	*dir = optional.Some("")
	return dir.Unwrap()
}

type GetMetaOptions struct {
	AbsPath        string
	CustomBuildCmd *string
	CustomStartCmd *string
	OutputDir      *string
}

func GetMeta(opt GetMetaOptions) PlanMeta {
	fs := afero.NewBasePathFs(afero.NewOsFs(), opt.AbsPath)

	packageJson, err := DeserializePackageJson(fs)
	if err != nil {
		log.Printf("Failed to read package.json: %v", err)
		// not fatal
	}

	ctx := &nodePlanContext{
		PackageJson: packageJson,
	}

	meta := PlanMeta{}

	framework := DetermineProjectFramework(ctx)
	meta["framework"] = string(framework)

	nodeVersion := GetNodeVersion(ctx)
	meta["nodeVersion"] = nodeVersion

	installCmd := GetInstallCmd(ctx, opt.AbsPath)
	meta["installCmd"] = installCmd

	buildCmd := GetBuildCmd(ctx, opt.AbsPath)
	if opt.CustomBuildCmd != nil && *opt.CustomBuildCmd != "" {
		buildCmd = *opt.CustomBuildCmd
	}
	meta["buildCmd"] = buildCmd

	if opt.OutputDir != nil && *opt.OutputDir != "" {
		if strings.HasPrefix(*opt.OutputDir, "/") {
			meta["outputDir"] = strings.TrimPrefix(*opt.OutputDir, "/")
		} else {
			meta["outputDir"] = *opt.OutputDir
		}
		return meta
	}
	staticOutputDir := GetStaticOutputDir(ctx)
	if staticOutputDir != "" {
		meta["outputDir"] = staticOutputDir
		return meta
	}

	startCmd := GetStartCmd(ctx, opt.AbsPath)
	if opt.CustomStartCmd != nil && *opt.CustomStartCmd != "" {
		startCmd = *opt.CustomStartCmd
	}
	meta["startCmd"] = startCmd

	return meta
}
