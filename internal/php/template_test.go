package php_test

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestTemplate(t *testing.T) {
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
		"nginx,owo",
	}
	property := []string{
		php.PropertyToString(types.PHPPropertyNone),
		php.PropertyToString(types.PHPPropertyComposer),
	}

	for _, v := range phpVersion {
		for _, d := range deps {
			for _, f := range framework {
				for _, p := range property {
					t.Run(v+"-"+f+"-"+d+"-"+p, func(t *testing.T) {
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
				}
			}
		}
	}
}

func TestTemplate_AcgFaka(t *testing.T) {
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
