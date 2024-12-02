package zeaburpack

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pan93412/envexpander/v3"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// InjectDockerfile injects the environment variables and
// the Docker.io registry into the Dockerfile.
func InjectDockerfile(dockerfile string, registry *string, variables map[string]string, planType types.PlanType, planMeta types.PlanMeta) string {
	// resolve env variable statically and don't depend on Dockerfile's order
	resolvedVars := envexpander.Expand(variables)

	refConstructor := newReferenceConstructor(registry)
	lines := strings.Split(dockerfile, "\n")
	stageLines := make([]int, 0)

	// construct the labels to indicate the language and framework used
	labels := []struct {
		Key   string
		Value string
	}{
		{
			Key:   "com.zeabur.zbpack.language",
			Value: string(planType),
		},
		{
			Key:   "com.zeabur.zbpack.framework",
			Value: planMeta["framework"],
		},
	}
	hasLanguageLabels := false

	for _, line := range lines {
		for _, label := range labels {
			if utils.WeakContains(line, label.Key) {
				hasLanguageLabels = true
				break
			}
		}
	}

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
		value := strconv.Quote(resolvedVars[key])
		dockerfileEnv += fmt.Sprintf(`ENV %s=%s`, key, value) + "\n"
	}

	for _, stageLine := range stageLines {
		lines[stageLine] = lines[stageLine] + "\n" + dockerfileEnv + "\n"
		if !hasLanguageLabels {
			for _, label := range labels {
				if label.Value == "" {
					continue
				}

				lines[stageLine] += fmt.Sprintf(`LABEL %s=%q`, label.Key, label.Value) + "\n"
			}
		}
	}

	return strings.Join(lines, "\n")
}
