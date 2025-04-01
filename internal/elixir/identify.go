package elixir

import (
	"github.com/spf13/afero"

	"github.com/salamer/zbpack/internal/utils"
	"github.com/salamer/zbpack/pkg/plan"
	"github.com/salamer/zbpack/pkg/types"
)

type identify struct{}

// NewIdentifier returns a new Elixir identifier.
func NewIdentifier() plan.Identifier {
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

var _ plan.Identifier = (*identify)(nil)
