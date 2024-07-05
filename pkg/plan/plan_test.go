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

func (mi alwaysMatchIdentifier) Explain(_ types.PlanMeta) []types.FieldInfo {
	return nil
}

func TestPlan_Continue(t *testing.T) {
	fs := afero.NewMemMapFs()

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "",
		},
		alwaysMatchIdentifier{plan.Continue()},
		alwaysMatchIdentifier{types.PlanMeta{"__INTERNAL_STATE": "TestPassed"}},
	)

	_, planMeta := planner.Plan()
	v, ok := planMeta["__INTERNAL_STATE"]

	assert.True(t, ok)
	assert.Equal(t, "TestPassed", v)
}

type identifierDemo struct{}

type explainableIdentifierDemo struct {
	identifierDemo
}

func (mi identifierDemo) PlanMeta(_ plan.NewPlannerOptions) types.PlanMeta {
	return types.PlanMeta{
		"framework":   "flutter",
		"dartVersion": "2.12.0",
	}
}

func (mi explainableIdentifierDemo) Explain(meta types.PlanMeta) []types.FieldInfo {
	return []types.FieldInfo{
		types.NewFrameworkFieldInfo("framework", types.PlanTypeDart, meta["framework"]),
		{
			Key:         "dartVersion",
			Name:        "Dart version",
			Description: "The version of Dart for building in the source code",
		},
	}
}

func (mi identifierDemo) PlanType() types.PlanType {
	return types.PlanTypeDart
}

func (mi identifierDemo) Match(_ afero.Fs) bool {
	return true
}

var _ plan.Identifier = (*explainableIdentifierDemo)(nil)

func TestPlan_FieldInfo(t *testing.T) {
	fs := afero.NewMemMapFs()

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "",
		},
		explainableIdentifierDemo{},
	)
	pt, pm := planner.Plan()

	explainer := plan.NewExplainer(explainableIdentifierDemo{})
	fieldInfo := explainer.Explain(pt, pm)

	assert.Len(t, fieldInfo, 3)
	assert.Equal(t, "Provider", fieldInfo[0].Name)
	assert.Equal(t, "Framework", fieldInfo[1].Name)
	assert.Equal(t, "Dart version", fieldInfo[2].Name)
}

func TestPlan_DefaultFieldInfo(t *testing.T) {
	fs := afero.NewMemMapFs()

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        fs,
			SubmoduleName: "",
		},
	)
	pt, pm := planner.Plan()

	explainer := plan.NewExplainer()
	fieldInfo := explainer.Explain(pt, pm)

	assert.Len(t, fieldInfo, 1)
	assert.Equal(t, "Provider", fieldInfo[0].Name)
}
