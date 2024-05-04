package dotnet

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Dotnet identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeDotnet
}

func (i *identify) Match(fs afero.Fs) bool {
	if utils.HasFile(fs, "Program.cs", "Startup.cs") {
		return true
	}

	fi, err := afero.ReadDir(fs, ".")
	if err == nil {
		return lo.ContainsBy(fi, func(f os.FileInfo) bool {
			return !f.IsDir() &&
				(strings.HasSuffix(f.Name(), ".sln") ||
					strings.HasSuffix(f.Name(), ".csproj"))
		})
	}

	return false
}

func (i *identify) findEntryPoint(
	fs afero.Fs,
	config plan.ImmutableProjectConfiguration,
	currentSubmoduleName string,
) (submoduleDir string, file string, err error) {
	moduleFs := fs

	if configSubmoduleDir, err := plan.Cast(
		config.Get("dotnet.submodule_dir"), cast.ToStringE,
	).Take(); err == nil && configSubmoduleDir != "" {
		submoduleDir = configSubmoduleDir
		moduleFs = afero.NewBasePathFs(fs, configSubmoduleDir)
	}

	if exist, _ := afero.Exists(moduleFs, currentSubmoduleName+".csproj"); exist {
		return submoduleDir, currentSubmoduleName + ".csproj", nil
	}

	files, err := afero.ReadDir(moduleFs, ".")
	if err != nil {
		return submoduleDir, "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".csproj") {
			return submoduleDir, file.Name(), nil
		}
	}

	return "", "", errors.New("no .csproj file found")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	submoduleDir, entryPoint, err := i.findEntryPoint(options.Source, options.Config, options.SubmoduleName)
	if err != nil {
		log.Printf("failed to find entrypoint: %s", err)
		return plan.Continue()
	}

	moduleFs := options.Source
	if submoduleDir != "" {
		moduleFs = afero.NewBasePathFs(options.Source, submoduleDir)
	}

	sdkVer, err := DetermineSDKVersion(entryPoint, moduleFs)
	if err != nil {
		log.Printf("failed to get sdk version: %s", err)
		return plan.Continue()
	}

	framework, err := DetermineFramework(entryPoint, moduleFs)
	if err != nil {
		log.Printf("failed to get framework: %s", err)
		return plan.Continue()
	}

	return types.PlanMeta{
		"sdk":          sdkVer,
		"entryPoint":   entryPoint,
		"submoduleDir": submoduleDir,
		"framework":    framework,
	}
}

var _ plan.Identifier = (*identify)(nil)
