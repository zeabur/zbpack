// Package dockerfile is the planner for projects already include Dockerfile.
package dockerfile

import (
	"fmt"
	"log"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"

	"github.com/zeabur/zbpack/pkg/types"
)

type dockerfilePlanContext struct {
	plan.NewPlannerOptions
}

// ReadDockerfile reads the Dockerfile in the project.
func ReadDockerfile(ctx *dockerfilePlanContext) ([]byte, error) {
	dockerfilePath, err := FindDockerfile(&FindContext{
		Source:        ctx.Source,
		Config:        ctx.Config,
		SubmoduleName: ctx.SubmoduleName,
	})
	if err != nil {
		return nil, fmt.Errorf("find dockerfile: %w", err)
	}

	content, err := utils.ReadFileToUTF8(ctx.Source, dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return content, nil
}

// GetMeta gets the meta of the Dockerfile project.
func GetMeta(opt plan.NewPlannerOptions) types.PlanMeta {
	ctx := &dockerfilePlanContext{
		NewPlannerOptions: opt,
	}

	dockerfileContent, err := ReadDockerfile(ctx)
	if err != nil {
		log.Printf("read dockerfile: %s", err)
		return plan.Continue()
	}

	meta := types.PlanMeta{
		"content": string(dockerfileContent),
	}
	return meta
}
