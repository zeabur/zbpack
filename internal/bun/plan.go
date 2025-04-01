package bun

import (
	"log"

	"github.com/moznion/go-optional"
	"github.com/salamer/zbpack/internal/nodejs"
	"github.com/salamer/zbpack/internal/utils"
	"github.com/salamer/zbpack/pkg/plan"
	"github.com/salamer/zbpack/pkg/types"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
)

// PlanContext is the context for Bun project planning.
type PlanContext struct {
	PackageJSON nodejs.PackageJSON
	Src         afero.Fs
	Config      plan.ImmutableProjectConfiguration

	Framework optional.Option[types.BunFramework]
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions nodejs.GetMetaOptions

// GetMeta gets the metadata of the Node.js project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := CreateBunContext(opt)

	meta := types.PlanMeta{}

	framework := DetermineFramework(ctx)
	meta["framework"] = string(framework)

	bunVersion := DetermineVersion(ctx)
	meta["bunVersion"] = bunVersion

	if framework != types.BunFrameworkNone {
		opt.BunFramework = optional.Some(framework)
	}

	meta = nodejs.GetMeta(nodejs.GetMetaOptions(opt))
	return meta
}

// CreateBunContext creates a new [PlanContext].
func CreateBunContext(opt GetMetaOptions) *PlanContext {
	packageJSON, err := nodejs.DeserializePackageJSON(opt.Src)
	if err != nil {
		log.Printf("Failed to read package.json: %v", err)
		// not fatal
	}

	return &PlanContext{
		PackageJSON: packageJSON,
		Src:         opt.Src,
		Config:      opt.Config,
	}
}

// DetermineFramework determines the framework of the Bun project.
func DetermineFramework(ctx *PlanContext) types.BunFramework {
	fw := &ctx.Framework
	packageJSON := ctx.PackageJSON

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	// Return None if Node.js framework is specified.
	if ctx.Config.Get("node.framework").IsSome() {
		*fw = optional.Some(types.BunFrameworkNone)
		return fw.Unwrap()
	}

	if framework, err := plan.Cast(ctx.Config.Get("bun.framework"), cast.ToStringE).Take(); err == nil {
		*fw = optional.Some(types.BunFramework(framework))
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

	*fw = optional.Some(types.BunFrameworkNone)
	return fw.Unwrap()
}

// DetermineVersion determines the Bun version to use.
func DetermineVersion(ctx *PlanContext) string {
	return utils.ConstraintToVersion(ctx.PackageJSON.Engines.Bun, "latest")
}
