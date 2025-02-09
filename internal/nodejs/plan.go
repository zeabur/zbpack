package nodejs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-yaml"
	"github.com/moznion/go-optional"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

const (
	// ConfigNodeFramework is the key for the configuration for specifying
	// the Node.js framework explicitly.
	//
	// This is an undocumented internal configuration and is subjected to change.
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

	PackageManager  optional.Option[PackageManager]
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

var (
	// NpmLatestMajorVersion is the latest major version of npm.
	NpmLatestMajorVersion uint64 = 10
	// NpmOldestMajorVersion is the oldest major version of npm.
	NpmOldestMajorVersion uint64 = 6
	// YarnLatestMajorVersions is the latest major version of yarn.
	YarnLatestMajorVersions uint64 = 4
	// YarnOldestMajorVersion is the oldest major version of yarn.
	YarnOldestMajorVersion uint64 = 1
	// PnpmLatestMajorVersion is the latest major version of pnpm.
	PnpmLatestMajorVersion uint64 = 10
	// PnpmOldestMajorVersion is the oldest major version of pnpm.
	PnpmOldestMajorVersion uint64 = 5
)

// packageManagerFieldRegex is the regular expression to match the package manager field in package.json.
// https://github.com/SchemaStore/schemastore/blob/d75f7a25e595611541644ca3051b1538f865504a/src/schemas/json/package.json#L739
var packageManagerFieldRegex = regexp.MustCompile(`(npm|pnpm|yarn|bun)@(\d+)\.\d+\.\d+(?:-.+)?`)

// DeterminePackageManager determines the package manager of the Node.js project.
func DeterminePackageManager(ctx *nodePlanContext) PackageManager {
	if pkgManager, err := ctx.PackageManager.Take(); err == nil {
		return pkgManager
	}

	pkgManager := DeterminePackageManagerUncached(ctx)
	ctx.PackageManager = optional.Some(pkgManager)
	return pkgManager
}

// DeterminePackageManagerUncached determines the package manager of the Node.js project.
func DeterminePackageManagerUncached(ctx *nodePlanContext) PackageManager {
	p := ctx.ProjectPackageJSON

	// Check packageManager.
	if p.PackageManager != nil && *p.PackageManager != "" {
		parsedPackageManager := packageManagerFieldRegex.FindStringSubmatch(*p.PackageManager)

		if len(parsedPackageManager) == 3 {
			switch parsedPackageManager[1] {
			case "npm":
				return Npm{MajorVersion: cast.ToUint64(parsedPackageManager[2])}
			case "pnpm":
				return Pnpm{MajorVersion: cast.ToUint64(parsedPackageManager[2])}
			case "yarn":
				return Yarn{MajorVersion: cast.ToUint64(parsedPackageManager[2])}
			case "bun":
				return Bun{}
			}
		}
	}

	// Check engines: https://github.com/nodejs/node/issues/51888
	if p.Engines.Bun != "" {
		return Bun{}
	}

	if p.Engines.Yarn != "" {
		return Yarn{MajorVersion: findContraintVersion(p.Engines.Yarn, YarnLatestMajorVersions, YarnOldestMajorVersion)}
	}

	if p.Engines.Pnpm != "" {
		return Pnpm{MajorVersion: findContraintVersion(p.Engines.Pnpm, PnpmLatestMajorVersion, PnpmOldestMajorVersion)}
	}

	if p.Engines.Npm != "" {
		return Npm{MajorVersion: findContraintVersion(p.Engines.Npm, NpmLatestMajorVersion, NpmOldestMajorVersion)}
	}

	// Check lockfiles.
	if utils.HasFile(ctx.Src, "yarn.lock") {
		return Yarn{}
	}

	if utils.HasFile(ctx.Src, "pnpm-lock.yaml") {
		return Pnpm{}
	}

	if utils.HasFile(ctx.Src, "bun.lock") || utils.HasFile(ctx.Src, "bun.lockb") {
		return Bun{}
	}

	if utils.HasFile(ctx.Src, "package-lock.json") {
		return Npm{}
	}

	return UnspecifiedPackageManager{PackageManager: Yarn{}}
}

func findContraintVersion(engineVersion string, latest uint64, oldest uint64) uint64 {
	// Try to parse engineVersion as a full semantic version.
	if v, err := semver.NewVersion(engineVersion); err == nil {
		major := v.Major()
		// If the parsed version's major is within the bounds, return it.
		if major >= oldest && major <= latest {
			return major
		}
	}

	// Otherwise, treat engineVersion as a constraint.
	constraint, err := semver.NewConstraint(engineVersion)
	if err != nil {
		return 0
	}

	// Iterate downward from the latest allowed major version.
	for i := latest; i >= oldest; i-- {
		v := semver.MustParse(fmt.Sprintf("%d.9999.9999", i))
		if constraint.Check(v) {
			return i
		}
	}

	return 0
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

	if _, isMedusa := packageJSON.Dependencies["@medusajs/medusa"]; isMedusa {
		*fw = optional.Some(types.NodeProjectFrameworkMedusa)
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

	if _, isMedusa := packageJSON.Dependencies["@medusajs/medusa"]; isMedusa {
		*fw = optional.Some(types.NodeProjectFrameworkMedusa)
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

// GetPredeployScript gets the predeploy script in package.json's `scripts` of the Node.js app.
func GetPredeployScript(ctx *nodePlanContext) string {
	packageJSON := ctx.GetAppPackageJSON()

	if _, ok := packageJSON.Scripts["predeploy"]; ok {
		return "predeploy"
	}

	return ""
}

// GetScriptCommand gets the command to run a script in the Node.js or Bun app.
func GetScriptCommand(ctx *nodePlanContext, script string) string {
	pkgManager := DeterminePackageManager(ctx)
	return pkgManager.GetRunScript(script)
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
	_, reldir := ctx.GetAppSource()

	if installCmd, err := cmd.Take(); err == nil {
		return installCmd
	}

	pkgManager := DeterminePackageManager(ctx)

	installCmdConf := plan.Cast(ctx.Config.Get(plan.ConfigInstallCommand), cast.ToStringE)

	var cmds []string

	needPlaywright := DetermineNeedPlaywright(ctx)
	if needPlaywright {
		cmds = append(
			cmds,
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdbus-1-3 libdrm2 libxkbcommon-x11-0 libxcomposite-dev libxdamage1 libxfixes-dev libxrandr2 libgbm-dev libasound2 && rm -rf /var/lib/apt/lists/*",
		)
	}

	needPuppeteer := DetermineNeedPuppeteer(ctx)
	if needPuppeteer {
		cmds = append(
			cmds,
			"RUN apt-get update && apt-get install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libgbm1 libasound2 libpangocairo-1.0-0 libxss1 libgtk-3-0 libxshmfence1 libglu1 && rm -rf /var/lib/apt/lists/*",
			"ENV PUPPETEER_CACHE_DIR=/src/.cache/puppeteer",
		)
	}

	if reldir != "" {
		cmds = append(cmds, "WORKDIR /src/"+reldir)
	}

	if installCmd, err := installCmdConf.Take(); err == nil {
		cmds = append(cmds, "COPY . .", "RUN "+installCmd)
	} else {
		initCommand := pkgManager.GetInitCommand()
		installDependenciesCommand := pkgManager.GetInstallProjectDependenciesCommand()
		cmds = append(cmds, "RUN "+initCommand, "COPY . .", "RUN "+installDependenciesCommand)
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

	if buildScript == "" {
		*cmd = optional.Some("")
		return cmd.Unwrap()
	}

	buildCmd := GetScriptCommand(ctx, buildScript)

	// if this is a Nitro-based framework, we should pass NITRO_PRESET
	// to the default build command.
	if slices.Contains(types.NitroBasedFrameworks, framework) {
		if serverless {
			buildCmd = "NITRO_PRESET=node " + buildCmd
		} else if pkgManager.GetType() == types.NodePackageManagerBun {
			buildCmd = "NITRO_PRESET=bun " + buildCmd
		} else {
			buildCmd = "NITRO_PRESET=node-server " + buildCmd
		}
	}

	if framework == types.NodeProjectFrameworkMedusa {
		installCmd := pkgManager.GetInstallProjectDependenciesCommand()

		// Install the dependencies in ".medusa/server" directory.
		buildCmd += " && " + "cd .medusa/server" + " && " + installCmd
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

	predeployScript := GetPredeployScript(ctx)

	startScript := GetStartScript(ctx)
	entry := GetEntry(ctx)
	framework := DetermineAppFramework(ctx)

	if startScript != "" {
		startCmd := GetScriptCommand(ctx, startScript)

		if predeployScript != "" {
			startCmd = GetScriptCommand(ctx, predeployScript) + " && " + startCmd
		}
		if framework == types.NodeProjectFrameworkMedusa {
			startCmd = "cd .medusa/server" + " && " + startCmd
		}

		*cmd = optional.Some(startCmd)
		return cmd.Unwrap()
	}

	var startCmd string
	runtime := lo.If(ctx.Bun, "bun").Else("node")

	if startScript == "" {
		switch {
		case entry != "":
			startCmd = runtime + " " + entry
		case framework == types.NodeProjectFrameworkSvelte:
			startCmd = runtime + " build/index.js"
		case types.IsNitroBasedFramework(string(framework)):
			startCmd = "HOST=0.0.0.0 " + runtime + " .output/server/index.mjs"
		default:
			startCmd = runtime + " index.js"
		}
	}

	if predeployScript != "" {
		predeployCommand := GetScriptCommand(ctx, predeployScript)

		*cmd = optional.Some(predeployCommand + " && " + startCmd)
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
	meta["packageManager"] = string(pkgManager.GetType())

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
