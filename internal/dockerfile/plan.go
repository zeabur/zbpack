// Package dockerfile is the planner for projects already include Dockerfile.
package dockerfile

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"golang.org/x/text/cases"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/zeabur/zbpack/pkg/types"
)

type dockerfilePlanContext struct {
	plan.NewPlannerOptions

	dockerfileName    optional.Option[string]
	dockerfileContent optional.Option[[]byte]
	ExposePort        optional.Option[string]
}

// ErrNoDockerfile is the error when there is no Dockerfile in the project.
var ErrNoDockerfile = errors.New("no dockerfile in this environment")

// ConfigDockerfileName is the key of the Dockerfile name in the config.
const ConfigDockerfileName = "dockerfile.name"

// FindDockerfile finds the Dockerfile in the project.
func FindDockerfile(ctx *dockerfilePlanContext) (string, error) {
	src := ctx.Source
	config := ctx.Config

	// Get the Dockerfile name from the config.
	// If there is not set, use the submodule as the Dockerfile name.
	dockerFilename := plan.Cast(
		config.Get("dockerfile.name"),
		cast.ToStringE,
	).TakeOr(ctx.SubmoduleName)

	path := &ctx.dockerfileName

	if path, err := ctx.dockerfileName.Take(); err == nil {
		return path, nil
	}

	dockerFilename, err := findDockerfile(src, dockerFilename)
	if err != nil {
		return "", err
	}

	*path = optional.Some(dockerFilename)
	return path.Unwrap(), nil
}

func findDockerfile(fs afero.Fs, dockerfileName string) (string, error) {
	converter := cases.Fold()

	files, err := afero.ReadDir(fs, ".")
	if err != nil {
		return "", err
	}

	foldedDockerfileName := converter.String(dockerfileName)

	// Create a map of all the files in the directory.
	// The filename here has been folded.
	type foldedFilename = string
	type originalFilename = string
	filesMap := make(map[foldedFilename]originalFilename, len(files))
	for _, file := range files {
		if file.Mode().IsRegular() {
			filesMap[converter.String(file.Name())] = file.Name()
		}
	}

	// Check if there is a Dockerfile.[submoduleName] or
	// [submoduleName].Dockerfile in the directory.
	// If there is, return it.
	if dockerfileName != "" {
		expectedFoldedFilename := "dockerfile." + foldedDockerfileName
		if originalFilename, ok := filesMap[expectedFoldedFilename]; ok {
			return originalFilename, nil
		}

		anotherExpectedFoldedFilename := foldedDockerfileName + ".dockerfile"
		if originalFilename, ok := filesMap[anotherExpectedFoldedFilename]; ok {
			return originalFilename, nil
		}
	}

	// Check if there is a Dockerfile in the directory.
	// If there is, return it.
	if originalFilename, ok := filesMap["dockerfile"]; ok {
		return originalFilename, nil
	}

	return "", ErrNoDockerfile
}

// ReadDockerfile reads the Dockerfile in the project.
func ReadDockerfile(ctx *dockerfilePlanContext) ([]byte, error) {
	c := &ctx.dockerfileContent

	if content, err := c.Take(); err == nil {
		return content, nil
	}

	dockerfileName, err := FindDockerfile(ctx)
	if err != nil {
		return nil, err
	}
	content, err := utils.ReadFileToUTF8(ctx.Source, dockerfileName)
	if err != nil {
		return nil, err
	}

	*c = optional.Some(content)
	return content, nil
}

// GetExposePort gets the exposed port of the Dockerfile project.
func GetExposePort(ctx *dockerfilePlanContext) string {
	const defaultValue = "8080"
	const exposePrefix = "EXPOSE "
	ctxPort := &ctx.ExposePort
	dockerFile, err := ReadDockerfile(ctx)
	if err != nil {
		return defaultValue
	}

	reader := bytes.NewReader(dockerFile)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.ToUpper(scanner.Text())
		line = strings.TrimSpace(line)

		port, found := strings.CutPrefix(line, exposePrefix)
		if !found {
			continue
		}
		if _, err := strconv.Atoi(port); err != nil {
			continue // not a valid `EXPOSE`
		}

		*ctxPort = optional.Some(port)
		return ctxPort.Unwrap()
	}

	*ctxPort = optional.Some(defaultValue)
	return defaultValue
}

// GetMeta gets the meta of the Dockerfile project.
func GetMeta(opt plan.NewPlannerOptions) types.PlanMeta {
	ctx := &dockerfilePlanContext{
		NewPlannerOptions: opt,
	}

	dockerfileContent, err := ReadDockerfile(ctx)
	if err != nil {
		log.Println(err)
		return plan.Continue()
	}

	exposePort := GetExposePort(ctx)

	meta := types.PlanMeta{
		"expose":  exposePort,
		"content": string(dockerfileContent),
	}
	return meta
}
