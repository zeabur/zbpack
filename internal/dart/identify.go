package dart

import (
	"strings"

	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

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

func determineFramework(source afero.Fs) types.DartFramework {
	file, err := afero.ReadFile(source, "pubspec.yaml")
	if err != nil {
		return types.DartFrameworkNone
	}

	if strings.Contains(string(file), "flutter") {
		return types.DartFrameworkFlutter
	}

	return types.DartFrameworkNone
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	framework := determineFramework(options.Source)

	meta := types.PlanMeta{
		"framework": string(framework),
	}

	if framework == types.DartFrameworkFlutter {
		meta["outputDir"] = "build/web"
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
