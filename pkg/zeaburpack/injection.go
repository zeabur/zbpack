package zeaburpack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pan93412/envexpander/v3"
	"github.com/zeabur/zbpack/pkg/types"
)

const (
	// DockerLabelLanguage is the label key for the language used in the Dockerfile.
	DockerLabelLanguage = "com.zeabur.zbpack.language"
	// DockerLabelFramework is the label key for the framework used in the Dockerfile.
	DockerLabelFramework = "com.zeabur.zbpack.framework"
)

// InjectDockerfile injects the environment variables and
// the Docker.io registry into the Dockerfile.
func InjectDockerfile(dockerfile string, registry *string, variables map[string]string, planType types.PlanType, planMeta types.PlanMeta) string {
	// resolve env variable statically and don't depend on Dockerfile's order
	resolvedVars := envexpander.Expand(variables)

	refConstructor := newReferenceConstructor(registry)
	lines := strings.Split(dockerfile, "\n")
	stageLines := make([]int, 0)

	labels := ExtractLabels(dockerfile)

	for i, line := range lines {
		fromStatement, isFromStatement := ParseFrom(line)
		if !isFromStatement {
			continue
		}

		// Construct the reference.
		newRef := refConstructor.Construct(fromStatement.Source)

		// Replace this FROM line.
		fromStatement.Source = newRef
		lines[i] = fromStatement.String()

		// Mark this FROM line as a stage.
		if stage, ok := fromStatement.Stage.Get(); ok {
			refConstructor.AddStage(stage)
		}

		stageLines = append(stageLines, i)
	}

	// sort the resolvedVars by key so we can build
	// the reproducible dockerfile
	sortedResolvedVarsKey := make([]string, 0, len(resolvedVars))
	for key := range resolvedVars {
		sortedResolvedVarsKey = append(sortedResolvedVarsKey, key)
	}
	sort.Strings(sortedResolvedVarsKey)

	// build the dockerfile
	dockerfileEnv := ""

	for _, key := range sortedResolvedVarsKey {
		dockerfileEnv += fmt.Sprintf(`ENV %s=%q`, key, resolvedVars[key]) + "\n"
	}

	for _, stageLine := range stageLines {
		lines[stageLine] = lines[stageLine] + "\n" + dockerfileEnv + "\n"

		// If this Dockerfile does not define the language, we can define it.
		if labels[DockerLabelLanguage] == "" {
			lines[stageLine] += fmt.Sprintf(
				`LABEL %s=%q %s=%q`,
				DockerLabelLanguage, string(planType),
				DockerLabelFramework, planMeta["framework"],
			) + "\n"
		}
	}

	return strings.Join(lines, "\n")
}

// UpdatePlanMetaWithLabel updates the plan type and plan meta
// with the language and framework labels.
func UpdatePlanMetaWithLabel(t types.PlanType, m types.PlanMeta, labels map[string]string) (types.PlanType, types.PlanMeta) {
	if language, ok := labels[DockerLabelLanguage]; ok {
		m["plannerLanguage"] = string(t) // save the original language
		t = types.PlanType(language)

		if framework, ok := labels[DockerLabelFramework]; ok {
			m["plannerFramework"] = m["framework"] // save the original framework
			m["framework"] = framework
		}
	}

	return t, m
}
