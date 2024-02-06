package php_test

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/php"
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
		"nginx",
		"nginx owo",
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

	for _, v := range phpVersion {
		v := v
		for _, d := range deps {
			d := d
			for _, f := range framework {
				f := f
				for _, p := range property {
					p := p
					t.Run(v+"-"+f+"-"+d+"-"+p, func(t *testing.T) {
						t.Parallel()

						dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
							"phpVersion": v,
							"framework":  f,
							"deps":       d,
							"app":        string(types.PHPApplicationDefault),
							"property":   p,
						})

						assert.NoError(t, err)
						snaps.MatchSnapshot(t, dockerfile)
					})

					if f == string(types.PHPFrameworkLaravel) {
						for _, o := range octaneServer {
							o := o

							t.Run(v+"-"+f+"-"+d+"-"+p+"+os-"+o, func(t *testing.T) {
								t.Parallel()

								dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
									"phpVersion":   v,
									"framework":    f,
									"deps":         d,
									"app":          string(types.PHPApplicationDefault),
									"property":     p,
									"octaneServer": o,
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

func TestTemplate_AcgFaka(t *testing.T) {
	t.Parallel()

	dockerfile, err := php.GenerateDockerfile(types.PlanMeta{
		"phpVersion": "8.2",
		"framework":  string(types.PHPFrameworkLaravel),
		"deps":       "nginx",
		"app":        string(types.PHPApplicationAcgFaka),
		"property":   php.PropertyToString(types.PHPPropertyComposer),
	})

	assert.NoError(t, err)
	snaps.MatchSnapshot(t, dockerfile)
}
