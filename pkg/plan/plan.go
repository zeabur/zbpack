// Package plan is the interface for planners.
package plan

import (
	"strconv"

	"github.com/spf13/afero"
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

	serverless := Cast(b.NewPlannerOptions.Config.Get("serverless"), ToWeakBoolE).TakeOr(true)

	return types.PlanTypeStatic, types.PlanMeta{"serverless": strconv.FormatBool(serverless)}
}
