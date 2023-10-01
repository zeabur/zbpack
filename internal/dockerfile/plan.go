// Package dockerfile is the planner for projects already include Dockerfile.
package dockerfile

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/pkg/types"
)

type dockerfilePlanContext struct {
	src afero.Fs

	dockerfileName    optional.Option[string]
	dockerfileContent optional.Option[[]byte]
	ExposePort        optional.Option[string]
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src           afero.Fs
	SubmoduleName string
}

// FindDockerfile finds the Dockerfile in the project.
func FindDockerfile(ctx *dockerfilePlanContext) (string, error) {
	src := ctx.src
	path := &ctx.dockerfileName

	if path, err := ctx.dockerfileName.Take(); err == nil {
		return path, nil
	}

	files, err := afero.ReadDir(src, ".")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.Mode().IsRegular() && strings.EqualFold(file.Name(), "Dockerfile") {
			*path = optional.Some(file.Name())
			return path.Unwrap(), nil
		}
	}

	return "", errors.New("no dockerfile in this environment")
}

// ReadDockerfile reads the Dockerfile in the project.
func ReadDockerfile(ctx *dockerfilePlanContext) ([]byte, error) {
	c := &ctx.dockerfileContent

	if content, err := c.Take(); err == nil {
		return []byte(content), nil
	}

	dockerfileName, err := FindDockerfile(ctx)
	if err != nil {
		return nil, err
	}
	content, err := afero.ReadFile(ctx.src, dockerfileName)
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
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := new(dockerfilePlanContext)
	ctx.src = opt.Src

	dockerfileContent, err := ReadDockerfile(ctx)
	if err != nil {
		log.Println(err)
		dockerfileContent = []byte{} // no Dockerfile
	}

	exposePort := GetExposePort(ctx)

	meta := types.PlanMeta{
		"expose":  exposePort,
		"content": string(dockerfileContent),
	}
	return meta
}
