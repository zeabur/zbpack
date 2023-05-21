package java

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Java identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeJava
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(
		fs, "pom.xml", "pom.yml", "pom.yaml", "build.gradle",
		"build.gradle.kts",
	)
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	projectType := DetermineProjectType(options.Source)
	framework := DetermineFramework(projectType, options.Source)
	jdkVersion := DetermineJDKVersion(projectType, options.Source)

	return types.PlanMeta{
		"type":      string(projectType),
		"framework": string(framework),
		"jdk":       jdkVersion,
	}
}

var _ plan.Identifier = (*identify)(nil)
