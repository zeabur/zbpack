package nodejs

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new NodeJS identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeNodejs
}

func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "package.json")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(
		GetMetaOptions{
			Src:            options.Source,
			Config:         options.Config,
			CustomBuildCmd: options.CustomBuildCommand,
			CustomStartCmd: options.CustomStartCommand,
			OutputDir:      options.OutputDir,
		},
	)
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	return Explain(meta)
}

// Explain returns the field information of the Node.js plan.
func Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		{
			Key:         "bun",
			Name:        "Enable Bun",
			Description: "Enable Bun as package manager.",
		},
		{
			Key:         "appDir",
			Name:        "Application Directory",
			Description: "The directory where the application to deploy is located.",
		},
		{
			Key:         "packageManager",
			Name:        "Package Manager",
			Description: "The package manager used to install the dependencies.",
		},
		types.NewFrameworkFieldInfo("framework", types.PlanTypeNodejs, meta["framework"]),
		{
			Key:         "nodeVersion",
			Name:        "Node.js version",
			Description: "The version of Node.js for building in the source code",
		},
		types.NewInstallCmdFieldInfo("installCmd"),
		types.NewBuildCmdFieldInfo("buildCmd"),
		types.NewStartCmdFieldInfo("startCmd"),
	}

	if _, ok := meta["serverless"]; ok {
		fieldInfo = append(fieldInfo, types.NewServerlessFieldInfo("serverless"))
	}

	if _, ok := meta["outputDir"]; ok {
		fieldInfo = append(fieldInfo, types.NewOutputDirFieldInfo("outputDir"))
	}

	return fieldInfo
}

var _ plan.ExplainableIdentifier = (*identify)(nil)
