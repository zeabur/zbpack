package java

import (
	"github.com/spf13/afero"
	"github.com/spf13/cast"

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
	targetExt := DetermineTargetExt(options.Source)

	planMeta := types.PlanMeta{
		"type":      string(projectType),
		"framework": string(framework),
		"targetExt": targetExt,
		"jdk":       jdkVersion,
	}

	javaArgs := plan.Cast(options.Config.Get("javaArgs"), cast.ToStringE)
	if args, err := javaArgs.Take(); err == nil {
		planMeta["javaArgs"] = args
	}

	return planMeta
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		{
			Key:         "type",
			Name:        "Build tool",
			Description: "The build tool used in the project.",
		},
		types.NewFrameworkFieldInfo("framework", types.PlanTypeJava, meta["framework"]),
		{
			Key:         "jdk",
			Name:        "JDK version",
			Description: "The version of JDK for building in the source code",
		},
	}

	if _, ok := meta["javaArgs"]; ok {
		fieldInfo = append(fieldInfo, types.FieldInfo{
			Key:         "javaArgs",
			Name:        "Java runtime arguments",
			Description: "The JVM arguments used for running the Java application.",
		})
	}

	// wip: targetExt, too verbose

	return fieldInfo
}

var _ plan.Identifier = (*identify)(nil)
