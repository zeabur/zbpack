package plan_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type alwaysMatchIdentifier struct {
	meta types.PlanMeta
}

func (mi alwaysMatchIdentifier) PlanType() types.PlanType {
	return ""
}

func (mi alwaysMatchIdentifier) Match(_ afero.Fs) bool {
	return true
}

func (mi alwaysMatchIdentifier) PlanMeta(_ plan.NewPlannerOptions) types.PlanMeta {
	return mi.meta
}

func TestPlan_Continue(t *testing.T) {
	fs := afero.NewMemMapFs()

	executor := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "",
		},
		alwaysMatchIdentifier{plan.Continue()},
		alwaysMatchIdentifier{types.PlanMeta{"__INTERNAL_STATE": "TestPassed"}},
	)

	_, planMeta := executor.Plan()
	v, ok := planMeta["__INTERNAL_STATE"]

	assert.True(t, ok)
	assert.Equal(t, "TestPassed", v)
}
