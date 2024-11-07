package dart

import (
	"strconv"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

type planContext struct {
	Config plan.ImmutableProjectConfiguration
	Src    afero.Fs

	Framework    optional.Option[types.DartFramework]
	BuildCommand optional.Option[string]
}

// NewIdentifier returns a new Ruby identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDart
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "pubspec.yaml")
}

func determineFramework(ctx planContext) types.DartFramework {
	src := ctx.Src
	f := &ctx.Framework

	if framework, err := f.Take(); err == nil {
		return framework
	}

	file, err := utils.ReadFileToUTF8(src, "pubspec.yaml")
	if err != nil {
		return types.DartFrameworkNone
	}

	if strings.Contains(string(file), "flutter") {
		*f = optional.Some(types.DartFrameworkFlutter)
		return f.Unwrap()
	}

	if strings.Contains(string(file), "serverpod") {
		*f = optional.Some(types.DartFrameworkServerpod)
		return f.Unwrap()
	}

	*f = optional.Some(types.DartFrameworkNone)
	return f.Unwrap()
}

func determineBuildCommand(ctx planContext) string {
	cfg := ctx.Config
	cmd := &ctx.BuildCommand

	if build, err := cmd.Take(); err == nil {
		return build
	}

	if customBuildCommand, err := plan.Cast(cfg.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		*cmd = optional.Some("RUN " + customBuildCommand)
		return cmd.Unwrap()
	}

	if determineFramework(ctx) == types.DartFrameworkFlutter {
		*cmd = optional.Some("RUN flutter build web")
		return cmd.Unwrap()
	}

	*cmd = optional.Some("RUN dart compile exe bin/main.dart")
	return cmd.Unwrap()
}

func determineOutputDir(ctx planContext) string {
	framework := determineFramework(ctx)

	if framework == types.DartFrameworkFlutter {
		return "build/web"
	}

	return ""
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	ctx := planContext{
		Src:    options.Source,
		Config: options.Config,
	}

	meta := types.PlanMeta{}

	if framework := determineFramework(ctx); framework != types.DartFrameworkNone {
		meta["framework"] = string(framework)
	}

	if build := determineBuildCommand(ctx); build != "" {
		meta["build"] = build
	}

	if od := determineOutputDir(ctx); od != "" {
		meta["serverless"] = strconv.FormatBool(
			utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(true),
		)
		meta["outputDir"] = od
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
