// Package plan is the interface for planners.
package plan

import (
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/pkg/types"
)

// Planner is the interface for planners.
type Planner interface {
	// Plan determines the information such as the language and framework of
	// the given project.
	Plan() (types.PlanType, types.PlanMeta)
}

// Explainer is the interface for explainers.
type Explainer interface {
	// Explain explains the given plan type and metadata.
	Explain(types.PlanType, types.PlanMeta) []types.FieldInfo
}

type planner struct {
	NewPlannerOptions

	identifiers []Identifier
}

type explainer struct {
	identifiers []Identifier
}

// NewPlannerOptions is the options for NewPlanner.
type NewPlannerOptions struct {
	Source             afero.Fs
	Config             ImmutableProjectConfiguration
	SubmoduleName      string
	CustomBuildCommand *string
	CustomStartCommand *string
	OutputDir          *string

	AWSConfig *AWSConfig
}

// AWSConfig is the AWS configuration for fetching projects from S3 bucket.
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// NewPlanner creates a new Planner.
func NewPlanner(opt *NewPlannerOptions, identifiers ...Identifier) Planner {
	return &planner{
		NewPlannerOptions: *opt,
		identifiers:       identifiers,
	}
}

// NewExplainer creates a new Explainer.
func NewExplainer(identifiers ...Identifier) Explainer {
	return &explainer{
		identifiers: identifiers,
	}
}

var continuePlanMeta = types.PlanMeta{
	"__INTERNAL_STATE": "CONTINUE",
}

// Continue is a pseudo PlanMeta, indicating the planner
// executor to find the next matched one.
func Continue() types.PlanMeta {
	return continuePlanMeta
}

func (b planner) Plan() (types.PlanType, types.PlanMeta) {
	for _, identifier := range b.identifiers {
		if identifier.Match(b.Source) {
			pt, pm := identifier.PlanType(), identifier.PlanMeta(b.NewPlannerOptions)

			// If the planner returns a Continue flag, we find the next matched.
			if v, ok := pm["__INTERNAL_STATE"]; ok && v == "CONTINUE" {
				continue
			}

			return pt, pm
		}
	}

	return types.PlanTypeStatic, types.PlanMeta{}
}

func (e explainer) Explain(planType types.PlanType, meta types.PlanMeta) []types.FieldInfo {
	identifier, found := lo.Find(e.identifiers, func(i Identifier) bool {
		return i.PlanType() == planType
	})
	if !found {
		return []types.FieldInfo{types.NewPlanTypeFieldInfo(planType)}
	}

	return addProviderToFieldInfo(identifier.Explain(meta), planType)
}

// addProviderToFieldInfo adds the plan type (as the key `_provider`) to the top of the field info.
func addProviderToFieldInfo(fieldInfo []types.FieldInfo, planType types.PlanType) []types.FieldInfo {
	newFieldInfo := make([]types.FieldInfo, len(fieldInfo)+1)

	newFieldInfo[0] = types.NewPlanTypeFieldInfo(planType)
	copy(newFieldInfo[1:], fieldInfo)

	return newFieldInfo
}
