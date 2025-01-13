package nodejs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
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

	// ConfigNodeFramework is the key for the configuration for specifying
	// the Node.js framework explicitly.
	ConfigNodeFramework = "node.framework"

	// ConfigAppDir indicates the relative path of the app to deploy.
	//
	// For example, if the app to deploy is located at `apps/api`,
	// the value of this configuration should be `apps/api`.
	ConfigAppDir = "app_dir"
)

type nodePlanContext struct {
	ProjectPackageJSON PackageJSON
	Config             plan.ImmutableProjectConfiguration
	Src                afero.Fs
	Bun                bool

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
	// AppDir is the directory of the application to deploy.
	AppDir optional.Option[string]
	// AppPackageJSON is the package.json of the app to deploy.
	AppPackageJSON optional.Option[PackageJSON]
}

// GetAppSource returns the source of the app to deploy of a Node.js project.
//
// A Node.js project may have a monorepo structure. In this case, the source
// of the app to deploy may not be the root; instead, it should be `apps/somewhere`.
//
// This function returns the real application directory and the relative path of application to project.
func (ctx *nodePlanContext) GetAppSource() (afero.Fs, string) {
	appDir := GetMonorepoAppRoot(ctx)
	if appDir == "" {
		return ctx.Src, ""
	}

	return afero.NewBasePathFs(ctx.Src, appDir), appDir
}

// GetAppPackageJSON returns the package.json of the app to deploy of a Node.js project.
func (ctx *nodePlanContext) GetAppPackageJSON() PackageJSON {
	if cachedPackageJSON, err := ctx.AppPackageJSON.Take(); err == nil {
		return cachedPackageJSON
	}

	src, relpath := ctx.GetAppSource()
	if relpath != "" {
		if packageJSON, err := DeserializePackageJSON(src); err == nil {
			ctx.AppPackageJSON = optional.Some(packageJSON)
			return packageJSON
		}
	}

	ctx.AppPackageJSON = optional.Some(ctx.ProjectPackageJSON)
	return ctx.AppPackageJSON.Unwrap()
}

// DeterminePackageManager determines the package manager of the Node.js project.
func DeterminePackageManager(ctx *nodePlanContext) types.NodePackageManager {
	src := ctx.Src
	pm := &ctx.PackageManager
	packageJSON := ctx.ProjectPackageJSON

	if packageManager, err := pm.Take(); err == nil {
		return packageManager
	}

	if ctx.Bun {
		*pm = optional.Some(types.NodePackageManagerBun)
		return pm.Unwrap()
	}

	if packageJSON.PackageManager != nil {
		// [pnpm]@8.4.0
		packageManagerSection := strings.SplitN(
			*packageJSON.PackageManager, "@", 2,
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

	if utils.HasFile(src, "bun.lockb") || utils.HasFile(src, "bun.lock") {
		*pm = optional.Some(types.NodePackageManagerBun)
		return pm.Unwrap()
	}

	*pm = optional.Some(types.NodePackageManagerUnknown)
	return pm.Unwrap()
}

// DetermineAppFramework determines the framework of the Node.js app.
func DetermineAppFramework(ctx *nodePlanContext) types.NodeProjectFramework {
	fw := &ctx.Framework
	packageJSON := ctx.GetAppPackageJSON()

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if framework, err := plan.Cast(ctx.Config.Get(ConfigNodeFramework), cast.ToStringE).Take(); err == nil {
		*fw = optional.Some(types.NodeProjectFramework(framework))
		return fw.Unwrap()
	}

	if _, isGrammY := packageJSON.Dependencies["grammy"]; isGrammY {
		*fw = optional.Some(types.NodeProjectFrameworkGrammY)
		return fw.Unwrap()
	}

	if _, isNuejs := packageJSON.Dependencies["nuejs-core"]; isNuejs {
		*fw = optional.Some(types.NodeProjectFrameworkNueJs)
		return fw.Unwrap()
	}

	if _, isAstro := packageJSON.FindDependency("astro"); isAstro {
		if _, hasZeaburAdapter := packageJSON.FindDependency("@zeabur/astro-adapter"); hasZeaburAdapter {
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
	if _, isSolid := packageJSON.FindDependency("@solidjs/start"); isSolid {
		*fw = optional.Some(types.NodeProjectFrameworkSolidStartVinxi)
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

	if _, isCreateReactApp := packageJSON.FindDependency("react-scripts"); isCreateReactApp {
		*fw = optional.Some(types.NodeProjectFrameworkCreateReactApp)
		return fw.Unwrap()
	}

	if _, isNuxtJs := packageJSON.FindDependency("nuxt"); isNuxtJs {
		*fw = optional.Some(types.NodeProjectFrameworkNuxtJs)
		return fw.Unwrap()
	}

	if _, isNitroPack := packageJSON.FindDependency("nitropack"); isNitroPack {
		*fw = optional.Some(types.NodeProjectFrameworkNitropack)
		return fw.Unwrap()
	}

	if _, isWaku := packageJSON.Dependencies["waku"]; isWaku {
		*fw = optional.Some(types.NodeProjectFrameworkWaku)
		return fw.Unwrap()
	}

	if _, isVitepress := packageJSON.FindDependency("vitepress"); isVitepress {
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

// DetermineNeedPuppeteer determines whether the app needs Puppeteer.
func DetermineNeedPuppeteer(ctx *nodePlanContext) bool {
	pup := &ctx.NeedPuppeteer
	appPackageJSON := ctx.GetAppPackageJSON()

	if needPuppeteer, err := pup.Take(); err == nil {
		return needPuppeteer
	}

	if _, hasPuppeteer := appPackageJSON.Dependencies["puppeteer"]; hasPuppeteer {
		*pup = optional.Some(true)
		return pup.Unwrap()
	}

	*pup = optional.Some(false)
	return pup.Unwrap()
}

// DetermineNeedPlaywright determines whether the app needs Playwright.
func DetermineNeedPlaywright(ctx *nodePlanContext) bool {
	pw := &ctx.NeedPlaywright
	appPackageJSON := ctx.GetAppPackageJSON()

	if needPlaywright, err := pw.Take(); err == nil {
		return needPlaywright
	}

	if _, hasPlaywright := appPackageJSON.Dependencies["playwright-chromium"]; hasPlaywright {
		*pw = optional.Some(true)
		return pw.Unwrap()
	}

	if _, hasPlaywright := appPackageJSON.DevDependencies["playwright-chromium"]; hasPlaywright {
		*pw = optional.Some(true)
		return pw.Unwrap()
	}

	*pw = optional.Some(false)
	return pw.Unwrap()
}

// GetBuildScript gets the build command in package.json's `scripts` of the Node.js app.
func GetBuildScript(ctx *nodePlanContext) string {
	bs := &ctx.BuildScript
	packageJSON := ctx.GetAppPackageJSON()

	if buildScript, err := bs.Take(); err == nil {
		return buildScript
	}

	if _, ok := packageJSON.Scripts["build"]; ok {
		*bs = optional.Some("build")
		return bs.Unwrap()
	}

	scriptsOrderedKey := make([]string, 0, len(packageJSON.Scripts))
	for key := range packageJSON.Scripts {
		scriptsOrderedKey = append(scriptsOrderedKey, key)
	}
	slices.Sort(scriptsOrderedKey)

	for _, key := range scriptsOrderedKey {
		if strings.Contains(key, "build") {
			*bs = optional.Some(key)
			return bs.Unwrap()
		}
	}

	*bs = optional.Some("")
	return bs.Unwrap()
}

// GetStartScript gets the start command in package.json's `scripts` of the Node.js app.
func GetStartScript(ctx *nodePlanContext) string {
	src, _ := ctx.GetAppSource()
	ss := &ctx.StartScript
	packageJSON := ctx.GetAppPackageJSON()

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

				if utils.HasFile(src, possibleEntrypoint) {
					*ss = optional.Some(possibleEntrypoint)
					return ss.Unwrap()
				}
			}
		}
	}

	*ss = optional.Some("")
	return ss.Unwrap()
}

const (
	defaultNodeVersion        = "20"
	maxNodeVersion     uint64 = 22
	maxLtsNodeVersion  uint64 = 20
)

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
	packageJSON := ctx.ProjectPackageJSON
	projectNodeVersion := packageJSON.Engines.Node

	// If there are ".node-version" or ".nvmrc" file, we pick
	// the version from them.
	if content, err := utils.ReadFileToUTF8(src, ".node-version"); err == nil {
		projectNodeVersion = strings.TrimSpace(string(content))
	}
	if content, err := utils.ReadFileToUTF8(src, ".nvmrc"); err == nil {
		projectNodeVersion = strings.TrimSpace(string(content))
	}

	return getNodeVersion(projectNodeVersion)
}

// GetEntry gets the entry file of the Node.js app.
func GetEntry(ctx *nodePlanContext) string {
	packageJSON := ctx.GetAppPackageJSON()
	ent := &ctx.Entry

	if entry, err := ent.Take(); err == nil {
		return entry
	}

	*ent = optional.Some(packageJSON.Main)
	return ent.Unwrap()
}

// GetInstallCmd gets the installation command of the Node.js app.
func GetInstallCmd(ctx *nodePlanContext) string {
	cmd := &ctx.InstallCmd
	src, reldir := ctx.GetAppSource()

	if installCmd, err := cmd.Take(); err == nil {
		return installCmd
	}

	pkgManager := DeterminePackageManager(ctx)

	// Disable cache_dependencies by default now due to some known cases:
	//
	//   * Monorepos: the critical dependencies are usually in the subdirectories.
	//   * Some postinstall scripts may require some files (other than package.json and
	//     lockfiles in the root)
	//   * Customized installation command
	//   * app root != project root (which means, there is more than 1 apps in this project)
	//
	// Considering we do not cache the Docker layer, let's disable it by default.
	shouldCacheDependencies := plan.Cast(ctx.Config.Get(ConfigCacheDependencies), plan.ToWeakBoolE).TakeOr(false)

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

	// disable cache_dependencies if the app root != project root
	if reldir != "" {
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
	if reldir != "" {
		cmds = append(cmds, "WORKDIR /src/"+reldir)
	}

	if installCmd, err := installCmdConf.Take(); err == nil {
		cmds = append(cmds, "RUN "+installCmd)
	} else {
		switch pkgManager {
		case types.NodePackageManagerNpm:
			// FIXME: reldir != ""
			if shouldCacheDependencies && reldir == "" {
				cmds = append(cmds, "COPY package-lock.json* .")
			}
			cmds = append(cmds, "RUN npm install")
		case types.NodePackageManagerPnpm:
			if shouldCacheDependencies && reldir == "" {
				cmds = append(cmds, "COPY pnpm-lock.yaml* .")
			}
			cmds = append(cmds, "RUN pnpm install")
		case types.NodePackageManagerBun:
			if shouldCacheDependencies && reldir == "" {
				cmds = append(cmds, "COPY bun.lock* .")
			}
			cmds = append(cmds, "RUN bun install")
		case types.NodePackageManagerYarn:
			if shouldCacheDependencies && reldir == "" {
				cmds = append(cmds, "COPY yarn.lock* .")
			}
			cmds = append(cmds, "RUN yarn install")
		default:
			cmds = append(cmds, "RUN yarn install")
		}
	}

	needPlaywright := DetermineNeedPlaywright(ctx)
	if needPlaywright {
		cmds = append([]string{
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdbus-1-3 libdrm2 libxkbcommon-x11-0 libxcomposite-dev libxdamage1 libxfixes-dev libxrandr2 libgbm-dev libasound2 && rm -rf /var/lib/apt/lists/*",
		}, cmds...)
	}

	needPuppeteer := DetermineNeedPuppeteer(ctx)
	if needPuppeteer {
		cmds = append([]string{
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1 && rm -rf /var/lib/apt/lists/*",
			"ENV PUPPETEER_CACHE_DIR=/src/.cache/puppeteer",
		}, cmds...)
	}

	*cmd = optional.Some(strings.Join(cmds, "\n"))
	return cmd.Unwrap()
}

// GetBuildCmd gets the build command of the Node.js app.
func GetBuildCmd(ctx *nodePlanContext) string {
	cmd := &ctx.BuildCmd

	if buildCmd, err := cmd.Take(); err == nil {
		return buildCmd
	}

	if buildCmd, err := plan.Cast(ctx.Config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		*cmd = optional.Some(buildCmd)
		return cmd.Unwrap()
	}

	buildScript := GetBuildScript(ctx)
	pkgManager := DeterminePackageManager(ctx)
	framework := DetermineAppFramework(ctx)
	serverless := getServerless(ctx)

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

	// if this is a Nitro-based framework, we should pass NITRO_PRESET
	// to the default build command.
	if slices.Contains(types.NitroBasedFrameworks, framework) {
		if serverless {
			buildCmd = "NITRO_PRESET=node " + buildCmd
		} else if pkgManager == types.NodePackageManagerBun {
			buildCmd = "NITRO_PRESET=bun " + buildCmd
		} else {
			buildCmd = "NITRO_PRESET=node-server " + buildCmd
		}
	}

	if buildScript == "" {
		buildCmd = ""
	}

	*cmd = optional.Some(buildCmd)
	return cmd.Unwrap()
}

// GetMonorepoAppRoot gets the app root of the monorepo project in the Node.js project.
func GetMonorepoAppRoot(ctx *nodePlanContext) string {
	if appDir, err := ctx.AppDir.Take(); err == nil {
		return appDir
	}

	// If user has explicitly set the app directory, we should use it.
	if userAppDir, err := plan.Cast(
		ctx.Config.Get(ConfigAppDir), cast.ToStringE,
	).Take(); err == nil && userAppDir != "" {
		if userAppDir == "/" {
			ctx.AppDir = optional.Some("")
			return ctx.AppDir.Unwrap()
		}

		ctx.AppDir = optional.Some(userAppDir)
		return ctx.AppDir.Unwrap()
	}

	// pnpm workspace
	workspace, found := func() (string, bool) {
		if workspaceYAML, err := afero.ReadFile(ctx.Src, "pnpm-workspace.yaml"); err == nil {
			var pnpmWorkspace struct {
				Packages []string `yaml:"packages"`
			}

			if err := yaml.Unmarshal(workspaceYAML, &pnpmWorkspace); err != nil {
				log.Printf("failed to parse pnpm-workspace.yaml: %v", err)
				return "", false
			}

			for _, pnpmPackagesGlob := range pnpmWorkspace.Packages {
				match, err := FindAppDirByGlob(ctx.Src, pnpmPackagesGlob)
				if err != nil {
					log.Printf("failed to find the matched directory: %v", err)
					continue
				}
				if match == "" {
					log.Printf("no directory found in the workspace according this glob: %s", pnpmPackagesGlob)
					continue
				}

				return match, true
			}

			return "", false
		}

		return "", false
	}()
	if found {
		ctx.AppDir = optional.Some(workspace)
		return ctx.AppDir.Unwrap()
	}

	// yarn workspace
	workspace, found = func() (string, bool) {
		if len(ctx.ProjectPackageJSON.Workspaces) == 0 {
			return "", false
		}

		for _, workspaceGlob := range ctx.ProjectPackageJSON.Workspaces {
			match, err := FindAppDirByGlob(ctx.Src, workspaceGlob)
			if err != nil {
				log.Printf("failed to find the matched directory: %v", err)
				continue
			}

			return match, true
		}

		return "", false
	}()
	if found {
		ctx.AppDir = optional.Some(workspace)
		return ctx.AppDir.Unwrap()
	}

	ctx.AppDir = optional.Some("")
	return ctx.AppDir.Unwrap()
}

// FindAppDirByGlob finds the application directory (with package.json) by the given glob pattern.
func FindAppDirByGlob(fs afero.Fs, pattern string) (match string, fnerr error) {
	matches, err := afero.Glob(fs, pattern)
	if err != nil {
		return "", err
	}

	for _, match := range matches {
		if _, err := DeserializePackageJSON(afero.NewBasePathFs(fs, match)); err != nil {
			fnerr = errors.Join(err, fmt.Errorf("deserialize package.json in %s: %w", match, err))
			continue
		}

		return match, nil
	}

	return "", fnerr
}

// GetStartCmd gets the start command of the Node.js app.
func GetStartCmd(ctx *nodePlanContext) string {
	cmd := &ctx.StartCmd

	if startCmd, err := cmd.Take(); err == nil {
		return startCmd
	}

	if getServerless(ctx) {
		*cmd = optional.Some("")
		return cmd.Unwrap()
	}

	// if the app is deployed as static files, we should not start the app.
	if GetStaticOutputDir(ctx) != "" {
		*cmd = optional.Some("")
		return cmd.Unwrap()
	}

	if startCmd, err := plan.Cast(ctx.Config.Get(plan.ConfigStartCommand), cast.ToStringE).Take(); err == nil {
		*cmd = optional.Some(startCmd)
		return cmd.Unwrap()
	}

	startScript := GetStartScript(ctx)
	pkgManager := DeterminePackageManager(ctx)
	entry := GetEntry(ctx)
	framework := DetermineAppFramework(ctx)

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
		switch {
		case entry != "":
			if ctx.Bun {
				startCmd = "bun " + entry
			} else {
				startCmd = "node " + entry
			}
		case framework == types.NodeProjectFrameworkSvelte:
			if ctx.Bun {
				startCmd = "bun build/index.js"
			} else {
				startCmd = "node build/index.js"
			}
		case types.IsNitroBasedFramework(string(framework)):
			if ctx.Bun {
				startCmd = "HOST=0.0.0.0 bun .output/server/index.mjs"
			} else {
				startCmd = "HOST=0.0.0.0 node .output/server/index.mjs"
			}
		default:
			if ctx.Bun {
				startCmd = "bun index.js"
			} else {
				startCmd = "node index.js"
			}
		}
	}

	// For solid-start projects, when using `solid-start start`
	// on solid-start-node, we should use the memory-efficient
	// start script instead.
	//
	// For more information, see the discussion in Discord: Solid.js
	// https://ptb.discord.com/channels/722131463138705510/
	// 722131463889223772/1140159307648868382
	if framework == types.NodeProjectFrameworkSolidStartNode && startScript == "start" {
		// solid-start-node specific start script
		startCmd = "node dist/server.js"
	}

	*cmd = optional.Some(startCmd)
	return cmd.Unwrap()
}

// GetStaticOutputDir returns the output directory for static application.
// If empty string is returned, the application is not deployed as static files.
func GetStaticOutputDir(ctx *nodePlanContext) string {
	dir := &ctx.StaticOutputDir
	source, _ := ctx.GetAppSource()

	if outputDir, err := dir.Take(); err == nil {
		return outputDir
	}

	if outputDir, err := plan.Cast(ctx.Config.Get(plan.ConfigOutputDir), cast.ToStringE).Take(); err == nil {
		*dir = optional.Some(outputDir)
		return dir.Unwrap()
	}

	framework := DetermineAppFramework(ctx)

	// the default output directory of Angular is `dist/<project-name>/browser`
	// we need to find the project name from `angular.json`.
	if framework == types.NodeProjectFrameworkAngular {
		angularJSON, err := utils.ReadFileToUTF8(source, "angular.json")
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

	// Vitepress's "build" script contains an additional parameter to specify
	// the output directory:
	//
	//     "build": "vitepress build [outdir]"
	//
	// We extract the outdir from the build command. If there is none,
	// we assume it is in the root directory of the project.
	if framework == types.NodeProjectFrameworkVitepress {
		buildScriptName := GetBuildScript(ctx)
		buildCommand := ctx.GetAppPackageJSON().Scripts[buildScriptName]

		// Extract the outdir from the build script.
		for _, buildCommandChunks := range strings.Split(buildCommand, "&&") {
			buildCommandChunks = strings.TrimSpace(buildCommandChunks)
			if outDir, ok := strings.CutPrefix(buildCommandChunks, "vitepress build"); ok {
				docsRoot := strings.TrimSpace(outDir)
				*dir = optional.Some(filepath.Join(docsRoot, ".vitepress", "dist"))
				return dir.Unwrap()
			}
		}
	}

	defaultStaticOutputDirs := map[types.NodeProjectFramework]string{
		types.NodeProjectFrameworkVite:             "dist",
		types.NodeProjectFrameworkUmi:              "dist",
		types.NodeProjectFrameworkVueCli:           "dist",
		types.NodeProjectFrameworkCreateReactApp:   "build",
		types.NodeProjectFrameworkHexo:             "public",
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
	sl := &ctx.Serverless

	if serverless, err := sl.Take(); err == nil {
		return serverless
	}

	if value, err := utils.GetExplicitServerlessConfig(ctx.Config).Take(); err == nil {
		*sl = optional.Some(value)
		return sl.Unwrap()
	}

	// For projects with outputDir, it should be always serverless (if not explicitly set).
	if GetStaticOutputDir(ctx) != "" {
		*sl = optional.Some(true)
		return sl.Unwrap()
	}

	// For monorepo projects, we should not deploy as serverless
	// until ZEA-3469 is resolved.
	if GetMonorepoAppRoot(ctx) != "" {
		*sl = optional.Some(false)
		return sl.Unwrap()
	}

	framework := DetermineAppFramework(ctx)

	defaultServerless := map[types.NodeProjectFramework]bool{
		types.NodeProjectFrameworkNextJs:  true,
		types.NodeProjectFrameworkAstro:   true,
		types.NodeProjectFrameworkSvelte:  true,
		types.NodeProjectFrameworkWaku:    true,
		types.NodeProjectFrameworkAngular: true,
		types.NodeProjectFrameworkRemix:   true,
	}
	for _, framework := range types.NitroBasedFrameworks {
		defaultServerless[framework] = true
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

	Bun          bool
	BunFramework optional.Option[types.BunFramework]
}

// GetMeta gets the metadata of the Node.js project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	packageJSON, err := DeserializePackageJSON(opt.Src)
	if err != nil {
		log.Printf("Failed to read package.json: %v", err)
		// not fatal
	}

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             opt.Config,
		Src:                opt.Src,
		Bun:                opt.Bun,
	}

	if bunFramework, err := opt.BunFramework.Take(); err == nil {
		// Bun and Node is interchangeable.
		ctx.Framework = optional.Some(types.NodeProjectFramework(bunFramework))
	}

	meta := types.PlanMeta{
		"bun": strconv.FormatBool(opt.Bun),
	}
	if opt.Bun {
		meta["bunVersion"] = "latest"
	}

	_, reldir := ctx.GetAppSource()
	meta["appDir"] = reldir

	pkgManager := DeterminePackageManager(ctx)
	meta["packageManager"] = string(pkgManager)

	framework := DetermineAppFramework(ctx)
	meta["framework"] = string(framework)

	nodeVersion := GetNodeVersion(ctx)
	meta["nodeVersion"] = nodeVersion

	installCmd := GetInstallCmd(ctx)
	meta["installCmd"] = installCmd

	buildCmd := GetBuildCmd(ctx)
	meta["buildCmd"] = buildCmd

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = strconv.FormatBool(serverless)
	}

	startCmd := GetStartCmd(ctx)
	meta["startCmd"] = startCmd

	// only set outputDir if there is no start command
	// (because if there is, it shouldn't be a static project)
	if startCmd == "" {
		staticOutputDir := GetStaticOutputDir(ctx)
		if staticOutputDir != "" {
			meta["outputDir"] = staticOutputDir
			return meta
		}
	}

	return meta
}
