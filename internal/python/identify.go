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
	return utils.HasFile(
		fs,
		"app.py", "main.py", "app.py", "manage.py", "requirements.txt", "streamlit_app.py",
	)
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	return GetMeta(GetMetaOptions{Src: options.Source, Config: options.Config})
}

var _ plan.Identifier = (*identify)(nil)
