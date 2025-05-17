package dockerfile

import (
	"log"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

const (
	// ConfigDockerfilePath is the configuration key for the Dockerfile path.
	ConfigDockerfilePath = "dockerfile.path"
	// ConfigDockerfileName is the configuration key for the Dockerfile name.
	// It equals to '/<name>.Dockerfile' (or '/Dockerfile.<name>')
	ConfigDockerfileName = "dockerfile.name"
)

type identify struct{}

// NewIdentifier returns a new Dockerfile identifier.
func NewIdentifier() plan.IdentifierV2 {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDocker
}

func (i *identify) Match(ctx plan.MatchContext) bool {
	fileInfo, err := afero.ReadDir(ctx.Source, ".")
	if err != nil {
		log.Println("dockerfile: read dir:", err)
		return false
	}

	dockerfilePath, err := plan.Cast(ctx.Config.Get(ConfigDockerfilePath), cast.ToStringE).Take()
	if err == nil && dockerfilePath != "" {
		_, err := ctx.Source.Stat(strings.Trim(dockerfilePath, "/"))
		if err == nil {
			return true
		}
	}

	dockerfileName := plan.Cast(ctx.Config.Get(ConfigDockerfileName), cast.ToStringE).TakeOr(ctx.SubmoduleName)

	// check if there is a 'Dockerfile' in the project.
	dockerfileNames := []string{
		"Dockerfile",
	}
	if dockerfileName != "" {
		dockerfileNames = append(dockerfileNames, "Dockerfile."+dockerfileName)
		dockerfileNames = append(dockerfileNames, dockerfileName+".Dockerfile")
	}

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		for _, name := range dockerfileNames {
			if strings.EqualFold(file.Name(), name) {
				return true
			}
		}
	}

	return false
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(options)
}
