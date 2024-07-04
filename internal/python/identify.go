package python

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Python identifier.
func NewIdentifier() plan.Identifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypePython
}

func (i *identify) Match(fs afero.Fs) bool {
	// some MkDocs projects may have requirements.txt, but it should be handled by static provider instead of Python
	if utils.HasFile(fs, "mkdocs.yml") {
		return false
	}

	return utils.HasFile(
		fs,
		"app.py", "main.py", "app.py", "manage.py", "requirements.txt",
		"streamlit_app.py", "pyproject.toml", "Pipfile",
	)
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(GetMetaOptions{Src: options.Source, Config: options.Config})
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		{
			Key:         "packageManager",
			Name:        "Package Manager",
			Description: "The package manager used to install the dependencies.",
		},
		{
			Key:         "pythonVersion",
			Name:        "Python Version",
			Description: "The version of Python for building in the source code",
		},
		types.NewInstallCmdFieldInfo("install"),
	}

	if _, ok := meta["framework"]; ok {
		fieldInfo = append(fieldInfo, types.NewBuildCmdFieldInfo("build"))
	}

	if _, ok := meta["build"]; ok {
		fieldInfo = append(fieldInfo, types.NewBuildCmdFieldInfo("build"))
	}

	if _, ok := meta["serverless"]; ok {
		fieldInfo = append(fieldInfo, types.NewServerlessFieldInfo("serverless"))
	}

	if _, ok := meta["entry"]; ok {
		fieldInfo = append(fieldInfo, types.FieldInfo{
			Key:         "entry",
			Name:        "WSGI entry",
			Description: "The WSGI entry point of the Python application",
		})
	}

	if _, ok := meta["start"]; ok {
		fieldInfo = append(fieldInfo, types.NewStartCmdFieldInfo("start"))
	}

	if _, ok := meta["selenium"]; ok {
		fieldInfo = append(fieldInfo, types.FieldInfo{
			Key:         "selenium",
			Name:        "Enable Selenium",
			Description: "Install Selenium and its dependencies in your application.",
		})
	}

	// WIP: static-flag; static-url-path; static-host-dir; apt-deps
	// they are so verbose and not necessary to present to users.

	return fieldInfo
}

var _ plan.Identifier = (*identify)(nil)
