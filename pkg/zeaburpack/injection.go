package zeaburpack

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pan93412/envexpander/v3"
)

// InjectDockerfile injects the environment variables and
// the Docker.io registry into the Dockerfile.
func InjectDockerfile(dockerfile string, registry *string, variables map[string]string) string {
	// resolve env variable statically and don't depend on Dockerfile's order
	resolvedVars := envexpander.Expand(variables)

	refConstructor := newReferenceConstructor(registry)
	lines := strings.Split(dockerfile, "\n")
	stageLines := make([]int, 0)

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
	}

	return strings.Join(lines, "\n")
}
