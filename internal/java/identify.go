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

// Match checks if the given filesystem contains files specific to Java projects.
func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(
		fs, "pom.xml", "pom.yml", "pom.yaml", "build.gradle",
		"build.gradle.kts",
	)
}

// PlanMeta returns metadata about the identified Java project.
func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	// Determine the project type
	projectType, err := DetermineProjectType(options.Source)
	if err != nil {
		projectType = "unknown" // Handle error gracefully, set default value
	}

	// Determine the framework
	framework, err := DetermineFramework(projectType, options.Source)
	if err != nil {
		framework = "unknown" // Handle error gracefully, set default value
	}

	// Determine JDK version
	jdkVersion, err := DetermineJDKVersion(projectType, options.Source)
	if err != nil {
		jdkVersion = "unknown" // Handle error gracefully, set default value
	}

	// Determine target extension
	targetExt := DetermineTargetExt(options.Source)

	// Create and return the PlanMeta
	return types.PlanMeta{
		"type":      string(projectType),
		"framework": string(framework),
		"targetExt": targetExt,
		"jdk":       jdkVersion,
	}
}

var _ plan.Identifier = (*identify)(nil)
