package ruby

import (
	"fmt"

	"github.com/zeabur/zbpack/pkg/types"
)

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	rubyVersion := meta["rubyVersion"]

	getRubyImage := fmt.Sprintf("FROM ruby:%s\n", rubyVersion)

	installCMD := `
RUN apt-get update -qq && apt-get install -y nodejs postgresql-client
`
	workDir := `
WORKDIR /myapp
`
	copyGemfile := `
COPY Gemfile /myapp/Gemfile
COPY Gemfile.lock /myapp/Gemfile.lock
`
	bundlerInstallCmd := `
RUN bundle install
`
	copySource := `
COPY . /myapp
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
