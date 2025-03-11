// Package plan is the interface for planners.
package plan

import (
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/pkg/types"
)

// Planner is the interface for planners.
type Planner interface {
	Plan() (types.PlanType, types.PlanMeta)
}

type planner struct {
	NewPlannerOptions

	identifiers []Identifier
}

// NewPlannerOptions is the options for NewPlanner.
type NewPlannerOptions struct {
	Source        afero.Fs
	Config        ImmutableProjectConfiguration
	SubmoduleName string

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

var continuePlanMeta = types.PlanMeta{
	"__INTERNAL_STATE": "CONTINUE",
}

// Continue is a pseudo PlanMeta, indicating the planner
// executor to find the next matched one.
func Continue() types.PlanMeta {
	return continuePlanMeta
}

const (
	// ConfigKeyPlanType is the key to specify plan type explicitly.
	// (ZBPACK_PLAN_TYPE)
	ConfigKeyPlanType = "plan_type"
)

func (b planner) Plan() (types.PlanType, types.PlanMeta) {
	planType, planTypeErr := Cast(b.NewPlannerOptions.Config.Get(ConfigKeyPlanType), cast.ToStringE).Take()

	if planTypeErr == nil {
		// find a identifier that matches the specified plan type
		identifier, ok := lo.Find(b.identifiers, func(i Identifier) bool {
			return i.PlanType() == types.PlanType(planType)
		})

		// if found, return the plan type and meta of this identifier
		if ok {
			return identifier.PlanType(), identifier.PlanMeta(b.NewPlannerOptions)
		}
	}

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
