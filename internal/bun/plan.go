package bun

import (
	"log"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

type bunPlanContext struct {
	PackageJSON nodejs.PackageJSON
	Src         afero.Fs

	Framework optional.Option[types.BunFramework]
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions nodejs.GetMetaOptions

// GetMeta gets the metadata of the Node.js project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	packageJSON, err := nodejs.DeserializePackageJSON(opt.Src)
	if err != nil {
		log.Printf("Failed to read package.json: %v", err)
		// not fatal
	}

	ctx := &bunPlanContext{
		PackageJSON: packageJSON,
		Src:         opt.Src,
	}

	meta := types.PlanMeta{}

	framework := DetermineFramework(ctx)
	meta["framework"] = string(framework)

	if framework == types.BunFrameworkHono {
		entry := determineEntry(ctx)
		if entry != "" {
			meta["entry"] = entry
		}

		return meta
	}

	meta = nodejs.GetMeta(nodejs.GetMetaOptions(opt))
	return meta
}

// DetermineFramework determines the framework of the Bun project.
func DetermineFramework(ctx *bunPlanContext) types.BunFramework {
	fw := &ctx.Framework
	packageJSON := ctx.PackageJSON

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	if _, isElysia := packageJSON.Dependencies["elysia"]; isElysia {
		*fw = optional.Some(types.BunFrameworkElysia)
		return fw.Unwrap()
	}

	if _, isBaojs := packageJSON.Dependencies["baojs"]; isBaojs {
		*fw = optional.Some(types.BunFrameworkBaojs)
		return fw.Unwrap()
	}

	if _, isBagel := packageJSON.Dependencies["@kakengloh/bagel"]; isBagel {
		*fw = optional.Some(types.BunFrameworkBagel)
		return fw.Unwrap()
	}

	if _, isHono := packageJSON.Dependencies["hono"]; isHono {
		*fw = optional.Some(types.BunFrameworkHono)
		return fw.Unwrap()
	}

	*fw = optional.Some(types.BunFrameworkNone)
	return fw.Unwrap()
}

func determineEntry(ctx *bunPlanContext) string {
	possibleEntries := []string{"index.ts", "index.js", "src/index.ts", "src/index.js"}

	for _, entry := range possibleEntries {
		if utils.HasFile(ctx.Src, entry) {
			return entry
		}
	}

	return ""
}
