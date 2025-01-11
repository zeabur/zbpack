package bun

import (
	"bytes"

	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Bun identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeBun
}

func (i *identify) Match(fs afero.Fs) bool {
	strategies := []func() bool{
		// has bun lockfile
		func() bool {
			return utils.HasFile(fs, "bun.lockb") || utils.HasFile(fs, "bun.lock")
		},
		// has .bun-version
		func() bool {
			return utils.HasFile(fs, ".bun-version")
		},
		// has bun types
		func() bool {
			packageJSON, err := utils.ReadFileToUTF8(fs, "package.json")
			if err != nil {
				return false
			}
			return bytes.Contains(packageJSON, []byte(`"bun-types"`))
		},
		// has bun@ (engine)
		func() bool {
			packageJSON, err := utils.ReadFileToUTF8(fs, "package.json")
			if err != nil {
				return false
			}
			return bytes.Contains(packageJSON, []byte(`bun@`))
		},
	}

	for _, strategy := range strategies {
		if strategy() {
			return true
		}
	}

	return false
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:    options.Source,
			Config: options.Config,
			Bun:    true,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
