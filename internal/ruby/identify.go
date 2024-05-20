package ruby

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Ruby identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeRuby
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "Gemfile")
}

// DetermineNeedNode determines if the project needs Node.js to build assets.
// This is a dirty hack because this should handle in Node.js provider.
func (i *identify) DetermineNeedNode(fs afero.Fs) bool {
	return utils.HasFile(fs, "package.json")
}

// DetermineNodePackageManager determines the Node.js package manager.
// This is a dirty hack because this should handle in Node.js provider.
func (i *identify) DetermineNodePackageManager(fs afero.Fs) types.NodePackageManager {
	if utils.HasFile(fs, "yarn.lock") {
		return types.NodePackageManagerYarn
	}

	if utils.HasFile(fs, "pnpm-lock.yaml") {
		return types.NodePackageManagerPnpm
	}

	return types.NodePackageManagerNpm
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	rubyVersion := DetermineRubyVersion(options.Source, options.Config)
	framework := DetermineRubyFramework(options.Source)
	buildCmd := DetermineBuildCmd(framework, options.Config)
	startCmd := DetermineStartCmd(framework, options.Config)

	meta := types.PlanMeta{
		"rubyVersion": rubyVersion,
		"buildCmd":    buildCmd,
		"startCmd":    startCmd,
	}

	needNode := i.DetermineNeedNode(options.Source)
	if needNode {
		meta["needNode"] = "true"
		meta["nodePackageManager"] = string(i.DetermineNodePackageManager(options.Source))
	}

	return meta
}

var _ plan.Identifier = (*identify)(nil)
