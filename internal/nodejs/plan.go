package nodejs

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

const (
	// ConfigCacheDependencies is the key for the configuration of
	// whether to cache dependencies.
	// It is true by default.
	ConfigCacheDependencies = "cache_dependencies"
)

type nodePlanContext struct {
	PackageJSON PackageJSON
	Config      plan.ImmutableProjectConfiguration
	Src         afero.Fs
	Bun         bool

	PackageManager  optional.Option[types.NodePackageManager]
	Framework       optional.Option[types.NodeProjectFramework]
	NeedPuppeteer   optional.Option[bool]
	NeedPlaywright  optional.Option[bool]
	BuildScript     optional.Option[string]
	StartScript     optional.Option[string]
	Entry           optional.Option[string]
	InstallCmd      optional.Option[string]
	BuildCmd        optional.Option[string]
	StartCmd        optional.Option[string]
	StaticOutputDir optional.Option[string]
	Serverless      optional.Option[bool]
}

// DeterminePackageManager determines the package manager of the Node.js project.
func DeterminePackageManager(ctx *nodePlanContext) types.NodePackageManager {
	src := ctx.Src
	pm := &ctx.PackageManager

	if packageManager, err := pm.Take(); err == nil {
		return packageManager
	}

	if ctx.Bun {
		*pm = optional.Some(types.NodePackageManagerBun)
		return pm.Unwrap()
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

	if utils.HasFile(src, "bun.lockb") {
		*pm = optional.Some(types.NodePackageManagerBun)
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

	if _, isGrammY := packageJSON.Dependencies["grammy"]; isGrammY {
		*fw = optional.Some(types.NodeProjectFrameworkGrammY)
		return fw.Unwrap()
	}

	if _, isNuejs := packageJSON.Dependencies["nuejs-core"]; isNuejs {
		*fw = optional.Some(types.NodeProjectFrameworkNueJs)
		return fw.Unwrap()
	}

	if _, isAstro := packageJSON.Dependencies["astro"]; isAstro {
		if _, hasZeaburAdapter := packageJSON.Dependencies["@zeabur/astro-adapter"]; hasZeaburAdapter {
			*fw = optional.Some(types.NodeProjectFrameworkAstro)
			return fw.Unwrap()
		}

		if _, isAstroSSR := packageJSON.Dependencies["@astrojs/node"]; isAstroSSR {
			*fw = optional.Some(types.NodeProjectFrameworkAstroSSR)
			return fw.Unwrap()
		}

		if _, isAstroStarlight := packageJSON.Dependencies["@astrojs/starlight"]; isAstroStarlight {
			*fw = optional.Some(types.NodeProjectFrameworkAstroStarlight)
			return fw.Unwrap()
		}

		*fw = optional.Some(types.NodeProjectFrameworkAstroStatic)
		return fw.Unwrap()
	}

	if _, isAstro := packageJSON.DevDependencies["astro"]; isAstro {
		*fw = optional.Some(types.NodeProjectFrameworkAstroStatic)
		return fw.Unwrap()
	}

	if _, isAngular := packageJSON.Dependencies["@angular/core"]; isAngular {
		*fw = optional.Some(types.NodeProjectFrameworkAngular)
		return fw.Unwrap()
	}

	if _, isSolid := packageJSON.FindDependency("solid-start"); isSolid {
		if _, isSolidStatic := packageJSON.FindDependency("solid-start-static"); isSolidStatic {
			*fw = optional.Some(types.NodeProjectFrameworkSolidStartStatic)
			return fw.Unwrap()
		}

		if _, isSolidNode := packageJSON.FindDependency("solid-start-node"); isSolidNode {
			*fw = optional.Some(types.NodeProjectFrameworkSolidStartNode)
			return fw.Unwrap()
		}

		*fw = optional.Some(types.NodeProjectFrameworkSolidStart)
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

	if _, isUmi := packageJSON.Dependencies["umi"]; isUmi {
		*fw = optional.Some(types.NodeProjectFrameworkUmi)
		return fw.Unwrap()
	}

	if _, isUmi := packageJSON.Dependencies["@umijs/max"]; isUmi {
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

	if _, isWaku := packageJSON.Dependencies["waku"]; isWaku {
		*fw = optional.Some(types.NodeProjectFrameworkWaku)
		return fw.Unwrap()
	}

	if _, isVitepress := packageJSON.DevDependencies["vitepress"]; isVitepress {
		*fw = optional.Some(types.NodeProjectFrameworkVitepress)
		return fw.Unwrap()
	}

	if _, isVueCliApp := packageJSON.DevDependencies["@vue/cli-service"]; isVueCliApp {
		*fw = optional.Some(types.NodeProjectFrameworkVueCli)
		return fw.Unwrap()
	}

	if _, isDocusaurus := packageJSON.Dependencies["@docusaurus/core"]; isDocusaurus {
		*fw = optional.Some(types.NodeProjectFrameworkDocusaurus)
		return fw.Unwrap()
	}

	if _, isVocs := packageJSON.Dependencies["vocs"]; isVocs {
		*fw = optional.Some(types.NodeProjectFrameworkVocs)
		return fw.Unwrap()
	}

	if _, isRspress := packageJSON.Dependencies["rspress"]; isRspress {
		*fw = optional.Some(types.NodeProjectFrameworkRspress)
		return fw.Unwrap()
	}

	if _, isVite := packageJSON.DevDependencies["vite"]; isVite {
		*fw = optional.Some(types.NodeProjectFrameworkVite)
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

// DetermineNeedPlaywright determines whether the project needs Playwright.
func DetermineNeedPlaywright(ctx *nodePlanContext) bool {
	pw := &ctx.NeedPlaywright
	packageJSON := ctx.PackageJSON

	if needPlaywright, err := pw.Take(); err == nil {
		return needPlaywright
	}

	if _, hasPlaywright := packageJSON.Dependencies["playwright-chromium"]; hasPlaywright {
		*pw = optional.Some(true)
		return pw.Unwrap()
	}

	if _, hasPlaywright := packageJSON.DevDependencies["playwright-chromium"]; hasPlaywright {
		*pw = optional.Some(true)
		return pw.Unwrap()
	}

	*pw = optional.Some(false)
	return pw.Unwrap()
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

	// If this is a Bun project, the entrypoint is usually
	// the value of `"module"`. Bun allows users to start the
	// application with `bun run <entrypoint>`.
	if ctx.Bun {
		if entrypoint := packageJSON.Module; entrypoint != "" {
			// The module path usually represents the artifact path
			// instead of the path where the source code is located.
			// We need to guess the correct extension of this entrypoint.

			// Remove the original trailing extension.
			finalDot := strings.LastIndex(entrypoint, ".")
			if finalDot != -1 {
				entrypoint = entrypoint[:finalDot]
			}

			// Find the possible entrypoint.
			for _, ext := range []string{
				".js", ".ts", ".tsx", ".jsx", ".mjs",
				".mts", ".cjs", ".cts",
			} {
				possibleEntrypoint := entrypoint + ext

				if utils.HasFile(ctx.Src, possibleEntrypoint) {
					*ss = optional.Some(possibleEntrypoint)
					return ss.Unwrap()
				}
			}
		}
	}

	*ss = optional.Some("")
	return ss.Unwrap()
}

const defaultNodeVersion = "18"
const maxNodeVersion uint64 = 21
const maxLtsNodeVersion uint64 = 18

func getNodeVersion(versionConstraint string) string {
	// .nvmrc extensions
	if versionConstraint == "node" {
		return strconv.FormatUint(maxNodeVersion, 10)
	}
	if versionConstraint == "lts/*" {
		return strconv.FormatUint(maxLtsNodeVersion, 10)
	}

	return utils.ConstraintToVersion(versionConstraint, defaultNodeVersion)
}

// GetNodeVersion gets the Node.js version of the project.
func GetNodeVersion(ctx *nodePlanContext) string {
	src := ctx.Src
	packageJSON := ctx.PackageJSON
	projectNodeVersion := packageJSON.Engines.Node

	// If there are ".node-version" or ".nvmrc" file, we pick
	// the version from them.
	if content, err := afero.ReadFile(src, ".node-version"); err == nil {
		projectNodeVersion = strings.TrimSpace(string(content))
	}
	if content, err := afero.ReadFile(src, ".nvmrc"); err == nil {
		projectNodeVersion = strings.TrimSpace(string(content))
	}

	return getNodeVersion(projectNodeVersion)
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

// GetInstallCmd gets the installation command of the Node.js project.
func GetInstallCmd(ctx *nodePlanContext) string {
	cmd := &ctx.InstallCmd
	src := ctx.Src

	if installCmd, err := cmd.Take(); err == nil {
		return installCmd
	}

	pkgManager := DeterminePackageManager(ctx)
	shouldCacheDependencies := plan.Cast(ctx.Config.Get(ConfigCacheDependencies), plan.ToWeakBoolE).TakeOr(true)

	// disable cache_dependencies for monorepos
	if shouldCacheDependencies && utils.HasFile(src, "pnpm-workspace.yaml", "pnpm-workspace.yml", "packages") {
		log.Println("Detected Monorepo. Disabling dependency caching.")
		shouldCacheDependencies = false
	}

	// disable cache_dependencies if the installation command is customized
	installCmdConf := plan.Cast(ctx.Config.Get(plan.ConfigInstallCommand), cast.ToStringE)
	if installCmdConf.IsSome() {
		shouldCacheDependencies = false
	}

	var cmds []string
	if shouldCacheDependencies {
		if utils.HasFile(src, "prisma") {
			cmds = append(cmds, "COPY prisma prisma")
		}
		cmds = append(cmds, "COPY package.json* tsconfig.json* .npmrc* .")
	} else {
		cmds = append(cmds, "COPY . .")
	}

	if installCmd, err := installCmdConf.Take(); err == nil {
		cmds = append(cmds, installCmd)
	} else {
		switch pkgManager {
		case types.NodePackageManagerNpm:
			cmds = append(cmds, "COPY package-lock.json* .", "RUN npm install")
		case types.NodePackageManagerPnpm:
			cmds = append(cmds, "COPY pnpm-lock.yaml* .", "RUN pnpm install")
		case types.NodePackageManagerBun:
			cmds = append(cmds, "COPY bun.lockb* .", "RUN bun install")
		case types.NodePackageManagerYarn:
			cmds = append(cmds, "COPY yarn.lock* .", "RUN yarn install")
		default:
			cmds = append(cmds, "RUN yarn install")
		}
	}

	needPlaywright := DetermineNeedPlaywright(ctx)
	if needPlaywright {
		cmds = append([]string{
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdbus-1-3 libdrm2 libxkbcommon-x11-0 libxcomposite-dev libxdamage1 libxfixes-dev libxrandr2 libgbm-dev libasound2",
		}, cmds...)
	}

	needPuppeteer := DetermineNeedPuppeteer(ctx)
	if needPuppeteer {
		cmds = append([]string{
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1",
			"ENV PUPPETEER_CACHE_DIR=/src/.cache/puppeteer",
		}, cmds...)
	}

	*cmd = optional.Some(strings.Join(cmds, "\n"))
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
	case types.NodePackageManagerBun:
		buildCmd = "bun run " + buildScript
	case types.NodePackageManagerYarn:
		fallthrough
	default:
		buildCmd = "yarn " + buildScript
	}

	if buildScript == "" {
		buildCmd = ""
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

	if getServerless(ctx) {
		*cmd = optional.Some("")
		return cmd.Unwrap()
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
	case types.NodePackageManagerBun:
		startCmd = "bun run " + startScript
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

	// For solid-start projects, when using `solid-start start`
	// on solid-start-node, we should use the memory-efficient
	// start script instead.
	//
	// For more information, see the discussion in Discord: Solid.js
	// https://ptb.discord.com/channels/722131463138705510/
	// 722131463889223772/1140159307648868382
	if framework == types.NodeProjectFrameworkSolidStartNode {
		if startScript == "start" {
			// solid-start-node specific start script
			startCmd = "node dist/server.js"
		}
	}

	*cmd = optional.Some(startCmd)
	return cmd.Unwrap()
}

// GetStaticOutputDir returns the output directory for static projects.
// If empty string is returned, the service is not deployed as static files.
func GetStaticOutputDir(ctx *nodePlanContext) string {
	dir := &ctx.StaticOutputDir
	source := ctx.Src

	if outputDir, err := dir.Take(); err == nil {
		return outputDir
	}

	framework := DetermineProjectFramework(ctx)

	// the default output directory of Angular is `dist/<project-name>/browser`
	// we need to find the project name from `angular.json`.
	if framework == types.NodeProjectFrameworkAngular {
		angularJSON, err := afero.ReadFile(source, "angular.json")
		if err != nil {
			println("failed to read angular.json: " + err.Error())
			*dir = optional.Some("dist")
			return dir.Unwrap()
		}

		type AngularJSON struct {
			Projects map[string]struct{} `json:"projects"`
		}

		var angular AngularJSON
		err = json.Unmarshal(angularJSON, &angular)
		if err != nil {
			println("failed to parse angular.json: " + err.Error())
			*dir = optional.Some("dist")
			return dir.Unwrap()
		}

		if len(angular.Projects) == 0 {
			println("no projects found in angular.json")
			*dir = optional.Some("dist")
			return dir.Unwrap()
		}

		var projectName string
		for name := range angular.Projects {
			projectName = name
			break
		}

		*dir = optional.Some("dist/" + projectName + "/browser")
		return dir.Unwrap()
	}

	defaultStaticOutputDirs := map[types.NodeProjectFramework]string{
		types.NodeProjectFrameworkVite:             "dist",
		types.NodeProjectFrameworkUmi:              "dist",
		types.NodeProjectFrameworkVueCli:           "dist",
		types.NodeProjectFrameworkCreateReactApp:   "build",
		types.NodeProjectFrameworkHexo:             "public",
		types.NodeProjectFrameworkVitepress:        "docs/.vitepress/dist",
		types.NodeProjectFrameworkAstroStatic:      "dist",
		types.NodeProjectFrameworkAstroStarlight:   "dist",
		types.NodeProjectFrameworkSliDev:           "dist",
		types.NodeProjectFrameworkDocusaurus:       "build",
		types.NodeProjectFrameworkSolidStartStatic: "dist/public",
		types.NodeProjectFrameworkVocs:             "docs/dist",
		types.NodeProjectFrameworkRspress:          "doc_build",
	}

	if outputDir, ok := defaultStaticOutputDirs[framework]; ok {
		*dir = optional.Some(outputDir)
		return dir.Unwrap()
	}

	*dir = optional.Some("")
	return dir.Unwrap()
}

func getServerless(ctx *nodePlanContext) bool {
	if value, err := utils.GetExplicitServerlessConfig(ctx.Config).Take(); err == nil {
		return value
	}

	sl := &ctx.Serverless

	if serverless, err := sl.Take(); err == nil {
		return serverless
	}

	framework := DetermineProjectFramework(ctx)

	defaultServerless := map[types.NodeProjectFramework]bool{
		types.NodeProjectFrameworkNextJs:  true,
		types.NodeProjectFrameworkNuxtJs:  true,
		types.NodeProjectFrameworkAstro:   true,
		types.NodeProjectFrameworkSvelte:  true,
		types.NodeProjectFrameworkWaku:    true,
		types.NodeProjectFrameworkAngular: true,
		types.NodeProjectFrameworkRemix:   true,
	}

	if serverless, ok := defaultServerless[framework]; ok {
		*sl = optional.Some(serverless)
		return sl.Unwrap()
	}

	*sl = optional.Some(false)
	return sl.Unwrap()
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration

	CustomBuildCmd *string
	CustomStartCmd *string
	OutputDir      *string

	Bun bool
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
		Config:      opt.Config,
		Src:         opt.Src,
		Bun:         opt.Bun,
	}

	meta := types.PlanMeta{
		"bun": strconv.FormatBool(opt.Bun),
	}

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

	// only set outputDir if there is no custom start command (because if there is, it shouldn't be a static project)
	if opt.CustomStartCmd == nil || *opt.CustomStartCmd == "" {

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

	}

	startCmd := GetStartCmd(ctx)
	if opt.CustomStartCmd != nil && *opt.CustomStartCmd != "" {
		startCmd = *opt.CustomStartCmd
	}
	meta["startCmd"] = startCmd

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = strconv.FormatBool(serverless)
	}

	return meta
}
