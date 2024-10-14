package golang

import (
	"os"
	"path"
	"strconv"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type goPlanContext struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration

	SubmoduleName string

	GoVersion optional.Option[string]
	Entry     optional.Option[string]

	Serverless optional.Option[bool]
}

const (
	// ConfigCgo indicates if cgo and its toolchains should be enabled.
	ConfigCgo = "go.cgo"
)

func getBuildCommand(ctx *goPlanContext) string {
	if buildCommand, err := plan.Cast(ctx.Config.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		return buildCommand
	}

	return ""
}

func isCgoEnabled(ctx *goPlanContext) bool {
	if cgo, err := plan.Cast(ctx.Config.Get(ConfigCgo), cast.ToBoolE).Take(); err == nil && cgo {
		return true
	}

	if os.Getenv("CGO_ENABLED") == "1" {
		return true
	}

	return false
}

func getEntry(ctx *goPlanContext) string {
	ent := &ctx.Entry
	if entry, err := ent.Take(); err == nil {
		return entry
	}

	// in a basic go project, we assume the entrypoint is main.go in root directory
	if utils.HasFile(ctx.Src, "main.go") {
		*ent = optional.Some("")
		return ent.Unwrap()
	}

	// if there is no main.go in root directory, we assume it's a monorepo project.
	// in a general monorepo Go repo of service "user-service", the entry point might be `./cmd/user-service/main.go`
	entry := path.Join("cmd", ctx.SubmoduleName, "main.go")
	if utils.HasFile(ctx.Src, entry) {
		*ent = optional.Some(entry)
		return ent.Unwrap()
	}

	// We know it's a Go project, but we don't know how to build it.
	// We'll just return a generic Go plan type.
	*ent = optional.Some("")
	return ""
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src           afero.Fs
	Config        plan.ImmutableProjectConfiguration
	SubmoduleName string
}

func getServerless(ctx *goPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
}

// GetMeta gets the metadata of the Go project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := &goPlanContext{
		Src:           opt.Src,
		Config:        opt.Config,
		SubmoduleName: opt.SubmoduleName,
	}
	meta := types.PlanMeta{}

	entry := getEntry(ctx)
	meta["entry"] = entry

	if build := getBuildCommand(ctx); build != "" {
		meta["build"] = build
	}

	meta["cgo"] = strconv.FormatBool(isCgoEnabled(ctx))

	serverless := getServerless(ctx)
	if serverless {
		meta["zeaburImageStage"] = "target-serverless"
	} else {
		meta["zeaburImageStage"] = "target-containerized"
	}

	return meta
}
