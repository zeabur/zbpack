package nodejs

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zeabur/zbpack/pkg/types"
)

// nitroServerPreset is the Nitro/Nuxt preset Zeabur actually runs.
// The official `zeabur` preset emits serverless `.zeabur/output`, which
// zbpack no longer deploys (PR #439).
func nitroServerPreset(ctx *nodePlanContext) string {
	if DeterminePackageManager(ctx).GetType() == types.NodePackageManagerBun {
		return "bun"
	}
	return "node-server"
}

// zeaburNitroPresetRe matches NITRO_PRESET=zeabur / SERVER_PRESET=zeabur
// (optional quotes) in build commands and package.json scripts.
var zeaburNitroPresetRe = regexp.MustCompile(`(?i)\b(?:NITRO_PRESET|SERVER_PRESET)\s*=\s*['"]?zeabur['"]?`)

// nitroPresetAssignRe matches any NITRO_PRESET / SERVER_PRESET assignment
// so the planner can replace them with the supported node-server/bun value.
var nitroPresetAssignRe = regexp.MustCompile(`(?i)\s*\b(?:NITRO_PRESET|SERVER_PRESET)\s*=\s*['"]?[\w.-]+['"]?`)

func packageScriptsHaveZeaburNitroPreset(pkg PackageJSON) bool {
	for _, script := range pkg.Scripts {
		if zeaburNitroPresetRe.MatchString(script) {
			return true
		}
	}
	return false
}

func rewriteZeaburNitroPresetInPackageJSON(preset string) string {
	// GNU sed on the official Node image. Unquoted assignments are what
	// the Nuxt/Nitro docs use; quoted forms still fail the output guard.
	return fmt.Sprintf("sed -i 's/NITRO_PRESET=zeabur/NITRO_PRESET=%s/g; s/SERVER_PRESET=zeabur/SERVER_PRESET=%s/g' package.json", preset, preset)
}

// applyNitroBuildOverrides remaps the unsupported official zeabur
// serverless preset to node-server (or bun) and forces those env vars
// for Nitro-based frameworks so a later assignment in the planner
// command cannot resurrect NITRO_PRESET=zeabur.
func applyNitroBuildOverrides(ctx *nodePlanContext, buildCmd string) string {
	if buildCmd == "" {
		return buildCmd
	}

	isNitro := slices.Contains(types.NitroBasedFrameworks, DetermineAppFramework(ctx))
	scriptsHaveZeabur := packageScriptsHaveZeaburNitroPreset(ctx.GetAppPackageJSON())
	if scriptsHaveZeabur || zeaburNitroPresetRe.MatchString(buildCmd) {
		isNitro = true
	}
	if !isNitro {
		return buildCmd
	}

	preset := nitroServerPreset(ctx)

	// Strip planner-level assignments first so prefixing the supported
	// preset is the value the build process actually sees.
	buildCmd = strings.TrimSpace(nitroPresetAssignRe.ReplaceAllString(buildCmd, " "))
	buildCmd = "NITRO_PRESET=" + preset + " SERVER_PRESET=" + preset + " " + buildCmd

	if scriptsHaveZeabur {
		// `pnpm build` re-runs the package.json script, whose inner
		// NITRO_PRESET=zeabur would otherwise override the prefix.
		buildCmd = rewriteZeaburNitroPresetInPackageJSON(preset) + " && " + buildCmd
	}

	return buildCmd
}
