package zeaburpack

import (
	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/pkg/types"
)

type generateDockerfileOptions struct {
	planType types.PlanType
	planMeta types.PlanMeta
}

func generateDockerfile(opt *generateDockerfileOptions) (string, error) {
	planType := opt.planType
	planMeta := opt.planMeta

	// find the packer
	for _, packer := range SupportedPackers() {
		if packer.PlanType() == planType {
			return packer.GenerateDockerfile(planMeta)
		}
	}

	// default to static
	dockerfile, err := static.GenerateDockerfile(planMeta)
	if err != nil {
		return "", err
	}

	return dockerfile, nil
}
