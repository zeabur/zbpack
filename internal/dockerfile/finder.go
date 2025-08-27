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
		trimmedPath := strings.Trim(dockerfilePath, "/")
		_, err := ctx.Source.Stat(trimmedPath)
		if err == nil {
			return trimmedPath, nil
		}
	}

	dockerfileName := plan.Cast(ctx.Config.Get(ConfigDockerfileName), cast.ToStringE).TakeOr(ctx.SubmoduleName)

	// check if there is a 'Dockerfile.[project-name]' in the project.
	fileInfo, err := afero.ReadDir(ctx.Source, ".")
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		if strings.EqualFold(file.Name(), "Dockerfile."+dockerfileName) || strings.EqualFold(file.Name(), dockerfileName+".Dockerfile") {
			return file.Name(), nil
		}
	}

	// check if there is a 'Dockerfile' in the project.
	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}
		if strings.EqualFold(file.Name(), "Dockerfile") {
			return file.Name(), nil
		}
	}

	return "", os.ErrNotExist
}
