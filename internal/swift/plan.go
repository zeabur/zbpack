package swift

import (
	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type swiftPlanContext struct {
	Src       afero.Fs
	Config    plan.ImmutableProjectConfiguration
	Framework optional.Option[types.SwiftFramework]
}

// DetermineFramework determines the framework of the Swift project.
func DetermineFramework(ctx *swiftPlanContext) types.SwiftFramework {
	src := ctx.Src
	fw := &ctx.Framework

	if framework, err := fw.Take(); err == nil {
		return framework
	}

	content, err := utils.ReadFileToUTF8(src, "Package.swift")
	if err != nil {
		*fw = optional.Some(types.SwiftFrameworkNone)
		return fw.Unwrap()
	}

	if utils.WeakContains(string(content), "https://github.com/vapor/vapor.git") {
		*fw = optional.Some(types.SwiftFrameworkVapor)
		return fw.Unwrap()
	}

	*fw = optional.Some(types.SwiftFrameworkNone)
	return fw.Unwrap()
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration
}

// GetMeta returns the metadata of a Swift project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	meta := types.PlanMeta{}

	ctx := &swiftPlanContext{
		Src:    opt.Src,
		Config: opt.Config,
	}

	framework := DetermineFramework(ctx)
	if framework != types.SwiftFrameworkNone {
		meta["framework"] = string(framework)
	}

	return meta
}
