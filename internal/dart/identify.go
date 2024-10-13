package dart

import (
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

	if framework, err := ctx.Framework.Take(); err == nil {
		return framework
	}

	file, err := utils.ReadFileToUTF8(src, "pubspec.yaml")
	if err != nil {
		return types.DartFrameworkNone
	}

	if strings.Contains(string(file), "flutter") {
		ctx.Framework = optional.Some(types.DartFrameworkFlutter)
		return ctx.Framework.Unwrap()
	}

	if strings.Contains(string(file), "serverpod") {
		ctx.Framework = optional.Some(types.DartFrameworkServerpod)
		return ctx.Framework.Unwrap()
	}

	ctx.Framework = optional.Some(types.DartFrameworkNone)
	return ctx.Framework.Unwrap()
}

func determineBuildCommand(ctx planContext) string {
	cfg := ctx.Config
	cmd := &ctx.BuildCommand

	if build, err := cmd.Take(); err == nil {
		return build
	}

	if customBuildCommand, err := plan.Cast(cfg.Get(plan.ConfigBuildCommand), cast.ToStringE).Take(); err == nil {
		*cmd = optional.Some(customBuildCommand)
		return cmd.Unwrap()
	}

	if determineFramework(ctx) == types.DartFrameworkFlutter {
		*cmd = optional.Some("flutter build web")
		return cmd.Unwrap()
	}

	*cmd = optional.Some("RUN dart compile exe bin/main.dart")
	return cmd.Unwrap()
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

	switch determineFramework(ctx) {
	case types.DartFrameworkFlutter:
		meta["zeaburImage"] = "dart-flutter"

		if utils.GetExplicitServerlessConfig(options.Config).TakeOr(true) {
			meta["zeaburImageStage"] = "target-static"
		} else {
			meta["zeaburImageStage"] = "target-containerized"
		}
	case types.DartFrameworkServerpod:
		meta["zeaburImage"] = "dart-serverpod"
	default:
		meta["zeaburImage"] = "dart-base"
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
