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
	hasPackageJSON := utils.HasFile(fs, "package.json")
	hasBunLockfile := utils.HasFile(fs, "bun.lockb")
	hasBunTypes := false

	packageJSON, err := utils.ReadFileToUTF8(fs, "package.json")
	if err == nil {
		hasBunTypes = bytes.Contains(packageJSON, []byte(`"bun-types"`))
	}

	// Some developer use bun as package manager for their Next.js or Nuxt.js project.
	// In this case, we should treat it as a Node.js project.
	if bytes.Contains(packageJSON, []byte(`"next"`)) || bytes.Contains(packageJSON, []byte(`"nuxt"`)) {
		return false
	}

	return hasPackageJSON && (hasBunLockfile || hasBunTypes)
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:            options.Source,
			Config:         options.Config,
			CustomBuildCmd: options.CustomBuildCommand,
			CustomStartCmd: options.CustomStartCommand,
			OutputDir:      options.OutputDir,
			Bun:            true,
		},
	)
}

var _ plan.Identifier = (*identify)(nil)
