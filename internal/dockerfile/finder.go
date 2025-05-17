package dockerfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/pkg/plan"
)

const (
	// ConfigDockerfilePath is the configuration key for the Dockerfile path.
	ConfigDockerfilePath = "dockerfile.path"
	// ConfigDockerfileName is the configuration key for the Dockerfile name.
	// It equals to '/<name>.Dockerfile' (or '/Dockerfile.<name>')
	ConfigDockerfileName = "dockerfile.name"
)

// FindContext is the context for finding the Dockerfile.
type FindContext struct {
	Source        afero.Fs
	Config        plan.ImmutableProjectConfiguration
	SubmoduleName string
}

// FindDockerfile returns the Dockerfile path we discovered.
// Return "os.ErrNotExist" if not found.
func FindDockerfile(ctx *FindContext) (filename string, err error) {
	dockerfilePath, err := plan.Cast(ctx.Config.Get(ConfigDockerfilePath), cast.ToStringE).Take()
	if err == nil && dockerfilePath != "" {
		_, err := ctx.Source.Stat(strings.Trim(dockerfilePath, "/"))
		if err == nil {
			return dockerfilePath, nil
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

	fileInfo, err := afero.ReadDir(ctx.Source, ".")
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		for _, name := range dockerfileNames {
			if strings.EqualFold(file.Name(), name) {
				return name, nil
			}
		}
	}

	return "", os.ErrNotExist
}
