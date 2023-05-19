package nodejs

import (
	"log"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

type nodePlanContext struct {
	PackageJSON PackageJSON
	Src         afero.Fs

	PackageManager  optional.Option[types.NodePackageManager]
	Framework       optional.Option[types.NodeProjectFramework]
	NeedPuppeteer   optional.Option[bool]
	BuildScript     optional.Option[string]
	StartScript     optional.Option[string]
	Entry           optional.Option[string]
	InstallCmd      optional.Option[string]
	BuildCmd        optional.Option[string]
	StartCmd        optional.Option[string]
	StaticOutputDir optional.Option[string]
}

// DeterminePackageManager determines the package manager of the Node.js project.
func DeterminePackageManager(ctx *nodePlanContext) types.NodePackageManager {
	src := ctx.Src
	pm := &ctx.PackageManager

	if packageManager, err := pm.Take(); err == nil {
		return packageManager
	}

	if ctx.PackageJSON.PackageManager != nil {
		// [pnpm]@8.4.0
		packageManagerSection := strings.SplitN(
			*ctx.PackageJSON.PackageManager, "@", 2,
		)

		switch packageManagerSection[0] {
		case "npm":
			*pm = optional.Some(types.NodePackageManagerNpm)
			return pm.Unwrap()
		case "yarn":
			*pm = optional.Some(types.NodePackageManagerYarn)
			return pm.Unwrap()
		case "pnpm":
			*pm = optional.Some(types.NodePackageManagerPnpm)
			return pm.Unwrap()
		default:
			log.Printf("Unknown package manager: %s", packageManagerSection[0])
			*pm = optional.Some(types.NodePackageManagerUnknown)
			return pm.Unwrap()
		}
	}

	if utils.HasFile(src, "yarn.lock") {
		*pm = optional.Some(types.NodePackageManagerYarn)
		return pm.Unwrap()
	}

	if utils.HasFile(src, "pnpm-lock.yaml") {
		*pm = optional.Some(types.NodePackageManagerPnpm)
		return pm.Unwrap()
	}

	if utils.HasFile(src, "package-lock.json") {
		*pm = optional.Some(types.NodePackageManagerNpm)
		return pm.Unwrap()
	}

	*pm = optional.Some(types.NodePackageManagerUnknown)
	return pm.Unwrap()
}

// DetermineProjectFramework determines the framework of the Node.js project.
func DetermineProjectFramework(ctx *nodePlanContext) types.NodeProjectFramework {
	fw := &ctx.Framework
	packageJSON := ctx.PackageJSON

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if _, isAstro := packageJSON.Dependencies["astro"]; isAstro {
		if _, isAstroSSR := packageJSON.Dependencies["@astrojs/node"]; isAstroSSR {
			*fw = optional.Some(types.NodeProjectFrameworkAstroSSR)
			return fw.Unwrap()
		}

		*fw = optional.Some(types.NodeProjectFrameworkAstroStatic)
		return fw.Unwrap()
	}

	if _, isSliDev := packageJSON.Dependencies["@slidev/cli"]; isSliDev {
		*fw = optional.Some(types.NodeProjectFrameworkSliDev)
		return fw.Unwrap()
	}

	if _, isSvelte := packageJSON.DevDependencies["svelte"]; isSvelte {
		*fw = optional.Some(types.NodeProjectFrameworkSvelte)
		return fw.Unwrap()
	}

	if _, isHexo := packageJSON.Dependencies["hexo"]; isHexo {
		*fw = optional.Some(types.NodeProjectFrameworkHexo)
		return fw.Unwrap()
	}

	if _, isQwik := packageJSON.DevDependencies["@builder.io/qwik"]; isQwik {
		*fw = optional.Some(types.NodeProjectFrameworkQwik)
		return fw.Unwrap()
	}

	if _, isVitepress := packageJSON.DevDependencies["vitepress"]; isVitepress {
		*fw = optional.Some(types.NodeProjectFrameworkVitepress)
		return fw.Unwrap()
	}

	if _, isVite := packageJSON.DevDependencies["vite"]; isVite {
		*fw = optional.Some(types.NodeProjectFrameworkVite)
		return fw.Unwrap()
	}

	if _, isUmi := packageJSON.Dependencies["umi"]; isUmi {
		*fw = optional.Some(types.NodeProjectFrameworkUmi)
		return fw.Unwrap()
	}

	if _, isNextJs := packageJSON.Dependencies["next"]; isNextJs {
		*fw = optional.Some(types.NodeProjectFrameworkNextJs)
		return fw.Unwrap()
	}

	if _, isNestJs := packageJSON.Dependencies["@nestjs/core"]; isNestJs {
		*fw = optional.Some(types.NodeProjectFrameworkNestJs)
		return fw.Unwrap()
	}

	if _, isRemix := packageJSON.Dependencies["@remix-run/react"]; isRemix {
		*fw = optional.Some(types.NodeProjectFrameworkRemix)
		return fw.Unwrap()
	}

	if _, isCreateReactApp := packageJSON.Dependencies["react-scripts"]; isCreateReactApp {
		*fw = optional.Some(types.NodeProjectFrameworkCreateReactApp)
		return fw.Unwrap()
	}

	if _, isNuxtJs := packageJSON.Dependencies["nuxt"]; isNuxtJs {
		*fw = optional.Some(types.NodeProjectFrameworkNuxtJs)
		return fw.Unwrap()
	}

	if _, isNuxtJs := packageJSON.DevDependencies["nuxt"]; isNuxtJs {
		*fw = optional.Some(types.NodeProjectFrameworkNuxtJs)
		return fw.Unwrap()
	}

	if _, isVueCliApp := packageJSON.DevDependencies["@vue/cli-service"]; isVueCliApp {
		*fw = optional.Some(types.NodeProjectFrameworkVueCli)
		return fw.Unwrap()
	}

	*fw = optional.Some(types.NodeProjectFrameworkNone)
	return fw.Unwrap()
}

// DetermineNeedPuppeteer determines whether the project needs Puppeteer.
func DetermineNeedPuppeteer(ctx *nodePlanContext) bool {
	pup := &ctx.NeedPuppeteer
	packageJSON := ctx.PackageJSON

	if needPuppeteer, err := pup.Take(); err == nil {
		return needPuppeteer
	}

	if _, hasPuppeteer := packageJSON.Dependencies["puppeteer"]; hasPuppeteer {
		*pup = optional.Some(true)
		return pup.Unwrap()
	}

	*pup = optional.Some(false)
	return pup.Unwrap()
}

// GetBuildScript gets the build command in package.json's `scripts` of the Node.js project.
func GetBuildScript(ctx *nodePlanContext) string {
	bs := &ctx.BuildScript
	packageJSON := ctx.PackageJSON

	if buildScript, err := bs.Take(); err == nil {
		return buildScript
	}

	if _, ok := packageJSON.Scripts["build"]; ok {
		*bs = optional.Some("build")
		return bs.Unwrap()
	}

	for key := range packageJSON.Scripts {
		if strings.Contains(key, "build") {
			*bs = optional.Some(key)
			return bs.Unwrap()
		}
	}

	*bs = optional.Some("")
	return bs.Unwrap()
}

// GetStartScript gets the start command in package.json's `scripts` of the Node.js project.
func GetStartScript(ctx *nodePlanContext) string {
	ss := &ctx.StartScript
	packageJSON := ctx.PackageJSON

	if startScript, err := ss.Take(); err == nil {
		return startScript
	}

	if _, ok := packageJSON.DevDependencies["@builder.io/qwik"]; ok {
		if _, ok := packageJSON.Scripts["deploy"]; ok {
			*ss = optional.Some("deploy")
			return ss.Unwrap()
		}
	}

	if _, ok := packageJSON.Scripts["start"]; ok {
		*ss = optional.Some("start")
		return ss.Unwrap()
	}

	*ss = optional.Some("")
	return ss.Unwrap()
}

const defaultNodeVersion = "16"

func getNodeVersion(versionRange string, versionsList []*semver.Version) string {
	if versionRange == "" {
		return defaultNodeVersion
	}

	// create a version constraint from versionRange
	constraint, err := semver.NewConstraint(versionRange)
	if err != nil {
		log.Println("invalid node version constraint", err)
		return defaultNodeVersion
	}

	// find the latest version which satisfies the constraint
	for _, version := range versionsList {
		if constraint.Check(version) {
			return strconv.FormatUint(version.Major(), 10)
		}
	}

	// when no version satisfies the constraint, return the default version
	return defaultNodeVersion
}

// GetNodeVersion gets the Node.js version of the project.
func GetNodeVersion(ctx *nodePlanContext) string {
	packageJSON := ctx.PackageJSON

	// nodeVersions is generated on compile time
	return getNodeVersion(packageJSON.Engines.Node, nodeVersions)
}

// GetEntry gets the entry file of the Node.js project.
func GetEntry(ctx *nodePlanContext) string {
	packageJSON := ctx.PackageJSON
	ent := &ctx.Entry

	if entry, err := ent.Take(); err == nil {
		return entry
	}

	*ent = optional.Some(packageJSON.Main)
	return ent.Unwrap()
}

// GetInstallCmd gets the install command of the Node.js project.
func GetInstallCmd(ctx *nodePlanContext) string {
	cmd := &ctx.InstallCmd

	if installCmd, err := cmd.Take(); err == nil {
		return installCmd
	}

	pkgManager := DeterminePackageManager(ctx)
	var installCmd string
	switch pkgManager {
	case types.NodePackageManagerNpm:
		installCmd = "npm install"
	case types.NodePackageManagerPnpm:
		installCmd = "pnpm install"
	case types.NodePackageManagerYarn:
		fallthrough
	default:
		installCmd = "yarn install"
	}

	*cmd = optional.Some(installCmd)
	return cmd.Unwrap()
}

// GetBuildCmd gets the build command of the Node.js project.
func GetBuildCmd(ctx *nodePlanContext) string {
	cmd := &ctx.BuildCmd

	if buildCmd, err := cmd.Take(); err == nil {
		return buildCmd
	}

	buildScript := GetBuildScript(ctx)
	pkgManager := DeterminePackageManager(ctx)

	var buildCmd string
	switch pkgManager {
	case types.NodePackageManagerPnpm:
		buildCmd = "pnpm run " + buildScript
	case types.NodePackageManagerNpm:
		buildCmd = "npm run " + buildScript
	case types.NodePackageManagerYarn:
		fallthrough
	default:
		buildCmd = "yarn " + buildScript
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

// GetStartCmd gets the start command of the Node.js project.
func GetStartCmd(ctx *nodePlanContext) string {
	cmd := &ctx.StartCmd

	if startCmd, err := cmd.Take(); err == nil {
		return startCmd
	}

	startScript := GetStartScript(ctx)
	pkgManager := DeterminePackageManager(ctx)
	entry := GetEntry(ctx)
	framework := DetermineProjectFramework(ctx)

	var startCmd string
	switch pkgManager {
	case types.NodePackageManagerPnpm:
		startCmd = "pnpm " + startScript
	case types.NodePackageManagerNpm:
		startCmd = "npm run " + startScript
	case types.NodePackageManagerYarn:
		fallthrough
	default:
		startCmd = "yarn " + startScript
	}

	if startScript == "" {
		if entry != "" {
			startCmd = "node " + entry
		} else if framework == types.NodeProjectFrameworkNuxtJs {
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

	defaultStaticOutputDirs := map[types.NodeProjectFramework]string{
		types.NodeProjectFrameworkVite:           "dist",
		types.NodeProjectFrameworkUmi:            "dist",
		types.NodeProjectFrameworkVueCli:         "dist",
		types.NodeProjectFrameworkCreateReactApp: "build",
		types.NodeProjectFrameworkHexo:           "public",
		types.NodeProjectFrameworkVitepress:      "docs/.vitepress/dist",
		types.NodeProjectFrameworkAstroStatic:    "dist",
		types.NodeProjectFrameworkSliDev:         "dist",
	}

	if outputDir, ok := defaultStaticOutputDirs[framework]; ok {
		*dir = optional.Some(outputDir)
		return dir.Unwrap()
	}

	*dir = optional.Some("")
	return dir.Unwrap()
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src            afero.Fs
	CustomBuildCmd *string
	CustomStartCmd *string
	OutputDir      *string
}

// GetMeta gets the metadata of the Node.js project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	packageJSON, err := DeserializePackageJSON(opt.Src)
	if err != nil {
		log.Printf("Failed to read package.json: %v", err)
		// not fatal
	}

	ctx := &nodePlanContext{
		PackageJSON: packageJSON,
		Src:         opt.Src,
	}

	meta := types.PlanMeta{}

	pkgManager := DeterminePackageManager(ctx)
	meta["packageManager"] = string(pkgManager)

	framework := DetermineProjectFramework(ctx)
	meta["framework"] = string(framework)

	nodeVersion := GetNodeVersion(ctx)
	meta["nodeVersion"] = nodeVersion

	installCmd := GetInstallCmd(ctx)
	meta["installCmd"] = installCmd

	buildCmd := GetBuildCmd(ctx)
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

	startCmd := GetStartCmd(ctx)
	if opt.CustomStartCmd != nil && *opt.CustomStartCmd != "" {
		startCmd = *opt.CustomStartCmd
	}
	meta["startCmd"] = startCmd

	return meta
}
