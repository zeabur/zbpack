// Package ruby is the build planner for Ruby projects.
package ruby

import (
	"fmt"
	"strings"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Ruby projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	rubyVersion := meta["rubyVersion"]

	getRubyImage := fmt.Sprintf("FROM docker.io/library/ruby:%s\n", rubyVersion)

	installSysDepCmd := []string{"RUN apt-get update -qq && apt-get install -y postgresql-client"}
	workDir := "WORKDIR /myapp"
	copySource := "COPY . /myapp"
	installDepCmd := []string{"RUN bundle install"}
	startCmd := "CMD " + meta["startCmd"]

	var precompileCmd string
	if buildCmd := meta["buildCmd"]; buildCmd != "" {
		precompileCmd = "RUN " + buildCmd
	}

	needNode := meta["needNode"] == "true"
	if needNode {
		installSysDepCmd = append(installSysDepCmd, "RUN apt-get install -y nodejs npm")

		switch meta["nodePackageManager"] {
		case "yarn":
			installSysDepCmd = append(installSysDepCmd, "RUN npm install -g yarn")
			installDepCmd = append(installDepCmd, "RUN yarn install")
		case "pnpm":
			installSysDepCmd = append(installSysDepCmd, "RUN npm install -g pnpm")
			installDepCmd = append(installDepCmd, "RUN pnpm install")
		default:
			installDepCmd = append(installDepCmd, "RUN npm install")
		}
	}

	dockerFile := getRubyImage + `
` + strings.Join(installSysDepCmd, "\n") + `
` + workDir + `
` + copySource + `
` + strings.Join(installDepCmd, "\n") + `
` + precompileCmd + `
` + startCmd

	return dockerFile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Ruby packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
