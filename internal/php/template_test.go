package php_test

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run `go-snaps` will sort snapshots
	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	phpVersion := []string{
		"8.1",
		"8.2",
		"7",
	}
	framework := []string{
		string(types.PHPFrameworkNone),
		string(types.PHPFrameworkLaravel),
		string(types.PHPFrameworkThinkphp),
		string(types.PHPFrameworkCodeigniter),
	}
	deps := []string{
		"",
		"nginx",
		"nginx owo",
	}
	exts := []string{
		"",
		"aaa",
		"aaa bbb",
	}
	property := []string{
		php.PropertyToString(types.PHPPropertyNone),
		php.PropertyToString(types.PHPPropertyComposer),
	}
	octaneServer := []string{
		"",
		"roadrunner",
		"swoole",
	}
	customStartCommands := []*string{
		nil,
		lo.ToPtr("php artisan serve"),
	}

	for _, v := range phpVersion {
		v := v
		for _, d := range deps {
			d := d
			for _, f := range framework {
				f := f
				for _, p := range property {
					p := p
					for _, e := range exts {
						e := e
						for _, customStartCommand := range customStartCommands {
							customStartCommand := customStartCommand

							t.Run(v+"-"+f+"-"+d+"-"+p+"-"+e+"-"+lo.FromPtrOr(customStartCommand, "none"), func(t *testing.T) {
								t.Parallel()

								config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")

								dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
									"phpVersion":   v,
									"framework":    f,
									"deps":         d,
									"exts":         e,
									"app":          string(types.PHPApplicationDefault),
									"property":     p,
									"startCommand": php.DetermineStartCommand(config, customStartCommand),
								})

								assert.NoError(t, err)
								snaps.MatchSnapshot(t, dockerfile)
							})

							if f == string(types.PHPFrameworkLaravel) {
								for _, o := range octaneServer {
									o := o

									t.Run(v+"-"+f+"-"+d+"-"+p+"-"+e+"+os-"+o+"-"+lo.FromPtrOr(customStartCommand, "none"), func(t *testing.T) {
										t.Parallel()

										config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")
										config.Set(php.ConfigLaravelOctaneServer, o)

										dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
											"phpVersion":   v,
											"framework":    f,
											"deps":         d,
											"exts":         e,
											"app":          string(types.PHPApplicationDefault),
											"property":     p,
											"octaneServer": o,
											"startCommand": php.DetermineStartCommand(config, customStartCommand),
										})

										assert.NoError(t, err)
										snaps.MatchSnapshot(t, dockerfile)
									})
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestTemplate_AcgFaka(t *testing.T) {
	t.Parallel()

	config := plan.NewProjectConfigurationFromFs(afero.NewMemMapFs(), "")

	dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
		"phpVersion":   "8.2",
		"framework":    string(types.PHPFrameworkLaravel),
		"deps":         "nginx",
		"app":          string(types.PHPApplicationAcgFaka),
		"property":     php.PropertyToString(types.PHPPropertyComposer),
		"startCommand": php.DetermineStartCommand(config, nil),
	})

	assert.NoError(t, err)
	snaps.MatchSnapshot(t, dockerfile)
}
