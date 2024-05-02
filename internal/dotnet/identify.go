package dotnet

import (
	"errors"
	"strings"

	"github.com/spf13/afero"

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
	return utils.HasFile(fs, "Program.cs", "Startup.cs")
}

func (i *identify) findEntryPoint(fs afero.Fs) (string, error) {
	files, err := afero.ReadDir(fs, ".")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".csproj") {
			return file.Name(), nil
		}
	}

	return "", errors.New("no .csproj file found")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	entryPoint, err := i.findEntryPoint(options.Source)
	if err != nil {
		panic(err)
	}

	sdkVer, err := DetermineSDKVersion(entryPoint, options.Source)
	if err != nil {
		panic(err)
	}

	framework, err := DetermineFramework(entryPoint, options.Source)
	if err != nil {
		panic(err)
	}

	return types.PlanMeta{
		"sdk":        sdkVer,
		"entryPoint": entryPoint,
		"framework":  framework,
	}
}

var _ plan.Identifier = (*identify)(nil)
