package nix

import (
	"bytes"

	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

const (
	// ConfigNixDockerPackage is the key for the package name of
	// `pkgs.dockerTools.buildImage` in the Nix configuration.
	//
	// For example, `packages.aarch64-linux.docker`.
	//
	// If not set, the default value is `packages.x86_64-linux.docker`.
	ConfigNixDockerPackage = "nix.docker_package"
)

// NewIdentifier returns a new Nix identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeNix
}

func (i *identify) Match(fs afero.Fs) bool {
	content, err := afero.ReadFile(fs, "flake.nix")
	if err != nil {
		return false
	}

	// <dockerTools>.[<buildImage>|<buildLayeredImage>|<streamLayeredImage>]
	return bytes.Contains(content, []byte("buildImage")) && bytes.Contains(content, []byte("dockerTools"))
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	packageName := plan.Cast(options.Config.Get(ConfigNixDockerPackage), cast.ToStringE).TakeOr("packages.x86_64-linux.docker")

	return types.PlanMeta{
		"package": packageName,
	}
}

var _ plan.Identifier = (*identify)(nil)
