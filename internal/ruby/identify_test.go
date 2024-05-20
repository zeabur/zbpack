package ruby_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run `go-snaps` will sort snapshots
	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestDockerfileSnapshotTest(t *testing.T) {
	t.Parallel()

	rubyVersions := []string{"2.7.2", "3.0.0"}
	needNodes := []string{"true", "false"}
	nodePackageManagers := []types.NodePackageManager{types.NodePackageManagerYarn, types.NodePackageManagerNpm, types.NodePackageManagerPnpm}
	buildCmds := []string{"", "bundle exec rake assets:precompile"}
	startCmds := []string{"ruby app.rb", "ruby main.rb", "rails server"}

	for _, rubyVersion := range rubyVersions {
		for _, needNode := range needNodes {
			for _, nodePackageManager := range nodePackageManagers {
				for _, buildCmd := range buildCmds {
					for _, startCmd := range startCmds {
						planMeta := types.PlanMeta{
							"rubyVersion":        rubyVersion,
							"needNode":           needNode,
							"nodePackageManager": string(nodePackageManager),
							"buildCmd":           buildCmd,
							"startCmd":           startCmd,
						}
						planMetaRepr := fmt.Sprintf("%#v", planMeta)

						t.Run(planMetaRepr, func(t *testing.T) {
							dockerfile, err := ruby.GenerateDockerfile(planMeta)
							if err != nil {
								t.Fatal(err)
							}

							snaps.MatchSnapshot(t, dockerfile)
						})
					}
				}
			}
		}
	}
}
