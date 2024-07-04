package elixir

import (
	"github.com/spf13/afero"

	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Elixir identifier.
func NewIdentifier() plan.ExplainableIdentifier {
	return &identify{}
}

func (i *identify) PlanType() types.PlanType {
	return types.PlanTypeElixir
}

// Match returns true if mix.exs is found in the source
func (i *identify) Match(fs afero.Fs) bool {
	return utils.HasFile(fs, "mix.exs")
}

func (i *identify) PlanMeta(options plan.NewPlannerOptions) types.PlanMeta {
	ElixirVer, err := DetermineElixirVersion(options.Source)
	if err != nil {
		panic(err)
	}

	ElixirFramework, err := DetermineElixirFramework(options.Source)
	if err != nil {
		panic(err)
	}

	ElixirEcto, err := CheckElixirEcto(options.Source)
	if err != nil {
		panic(err)
	}

	return types.PlanMeta{
		"ver":       ElixirVer,
		"framework": ElixirFramework,
		"ecto":      ElixirEcto,
	}
}

func (i *identify) Explain(meta types.PlanMeta) []types.FieldInfo {
	fieldInfo := []types.FieldInfo{
		{
			Key:         "ver",
			Name:        "Elixir Version",
			Description: "The version of Elixir for building in the source code",
		},
		types.NewFrameworkFieldInfo("framework", types.PlanTypeElixir, meta["framework"]),
		{
			Key:         "ecto",
			Name:        "Ecto Project",
			Description: "Is this project using Ecto?",
		},
	}

	return fieldInfo
}

var _ plan.ExplainableIdentifier = (*identify)(nil)
