package ruby

import (
	"fmt"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Ruby projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	rubyVersion := meta["rubyVersion"]

	getRubyImage := fmt.Sprintf("FROM docker.io/library/ruby:%s\n", rubyVersion)

	// ROR framework requires nodejs and postgresql-client
	installCMD := `
RUN apt-get update -qq && apt-get install -y nodejs postgresql-client
`
	workDir := `
WORKDIR /myapp
`
	// copy gemfile for install package
	copyGemfile := `
COPY Gemfile /myapp/Gemfile
COPY Gemfile.lock /myapp/Gemfile.lock
`
	bundlerInstallCmd := `
RUN bundle install
`
	// copy source to workdir
	copySource := `
COPY . /myapp
RUN bundle exec rake assets:precompile
`
	startCmd := `
EXPOSE ${PORT}
CMD ["rails", "server", "-b", "0.0.0.0","-p","8080"]
`
	dockerFile := getRubyImage +
		installCMD +
		workDir +
		copyGemfile +
		bundlerInstallCmd +
		copySource +
		startCmd

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
