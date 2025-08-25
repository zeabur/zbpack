package zeaburpack

import (
	"fmt"
	"strings"

	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/pkg/types"
)

type generateDockerfileOptions struct {
	planType types.PlanType
	planMeta types.PlanMeta
}

// injectLabels injects language and framework labels into the Dockerfile.
// It adds the labels after the first FROM statement in the Dockerfile.
func injectLabels(dockerfile string, planType types.PlanType, planMeta types.PlanMeta) string {
	lines := strings.Split(dockerfile, "\n")
	var injectedLines []string
	labelsInjected := false

	for _, line := range lines {
		injectedLines = append(injectedLines, line)

		// Look for the first FROM statement and inject labels after it
		if !labelsInjected && strings.HasPrefix(strings.TrimSpace(strings.ToUpper(line)), "FROM ") {
			// Add language label
			languageLabel := fmt.Sprintf(`LABEL "language"="%s"`, planType)
			injectedLines = append(injectedLines, languageLabel)

			// Add framework label if framework exists in meta
			if framework, exists := planMeta["framework"]; exists && framework != "" {
				frameworkLabel := fmt.Sprintf(`LABEL "framework"="%s"`, framework)
				injectedLines = append(injectedLines, frameworkLabel)
			}

			labelsInjected = true
		}
	}

	// If no FROM statement was found, add labels at the beginning
	if !labelsInjected {
		var labelsToAdd []string
		languageLabel := fmt.Sprintf(`LABEL "language"="%s"`, planType)
		labelsToAdd = append(labelsToAdd, languageLabel)

		if framework, exists := planMeta["framework"]; exists && framework != "" {
			frameworkLabel := fmt.Sprintf(`LABEL "framework"="%s"`, framework)
			labelsToAdd = append(labelsToAdd, frameworkLabel)
		}

		// Insert labels at the beginning
		injectedLines = append(labelsToAdd, injectedLines...)
	}

	return strings.Join(injectedLines, "\n")
}

func generateDockerfile(opt *generateDockerfileOptions) (string, error) {
	planType := opt.planType
	planMeta := opt.planMeta

	var dockerfile string
	var err error

	// find the packer
	found := false
	for _, packer := range SupportedPackers() {
		if packer.PlanType() == planType {
			dockerfile, err = packer.GenerateDockerfile(planMeta)
			found = true
			break
		}
	}

	if !found {
		// default to static
		dockerfile, err = static.GenerateDockerfile(planMeta)
	}

	if err != nil {
		return "", err
	}

	// Inject language and framework labels
	dockerfile = injectLabels(dockerfile, planType, planMeta)

	return dockerfile, nil
}
