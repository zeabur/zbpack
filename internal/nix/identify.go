package nix

import (
	"log"
	"regexp"
	"runtime"
	"strings"

	"github.com/samber/lo"
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
	// If not set, the default value is
	// `packages.<arch>-linux.{dockerImage|docker-image|docker|image|container}`.
	// where the first found package will be used.
	ConfigNixDockerPackage = "nix.docker_package"
)

var defaultDockerPackageFinderRegex = regexp.MustCompile(`(?i)(?m)^.*?(dockerImage|docker-image|docker|image|container)\s*=.+$`)

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

	// <dockerTools>.[<buildImage>|<buildLayeredImage>]
	return strings.Contains(string(content), "dockerTools") && (strings.Contains(string(content), "buildImage") || strings.Contains(string(content), "buildLayeredImage"))
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	arch := lo.If(runtime.GOARCH == "amd64", "x86_64").ElseIf(runtime.GOARCH == "arm64", "aarch64").Else("x86_64")

	packageName, err := plan.Cast(options.Config.Get(ConfigNixDockerPackage), cast.ToStringE).Take()
	if err != nil {
		content, err := afero.ReadFile(options.Source, "flake.nix")
		if err != nil {
			return plan.Continue()
		}

		packageName = FindPossibleNixDockerPackage(string(content))
		if packageName == "" {
			log.Println("warning: no default package found; skipping Nix planner")
			return plan.Continue()
		}

		// Add x86_64 prefix :D
		packageName = "packages." + arch + "-linux." + packageName
	}

	return types.PlanMeta{
		"package": packageName,
	}
}

// FindPossibleNixDockerPackage finds the possible Nix package name for Docker.
func FindPossibleNixDockerPackage(content string) string {
	group := defaultDockerPackageFinderRegex.FindStringSubmatch(content)
	if len(group) != 2 {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(group[0]), "#") {
		return ""
	}

	return strings.TrimSpace(group[1])
}

var _ plan.Identifier = (*identify)(nil)
