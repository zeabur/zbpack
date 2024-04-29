package utils

import (
	"github.com/moznion/go-optional"
	"github.com/zeabur/zbpack/pkg/plan"
)

// GetExplicitServerlessConfig gets the serverless flag from the project configuration.
//
// When the serverless flag is not set, it will be determined by the FORCE_CONTAINERIZED
// and ZBPACK_SERVERLESS environment variables.
// If all of them are not set, it returns None for consumers to determine the default value.
func GetExplicitServerlessConfig(config plan.ImmutableProjectConfiguration) optional.Option[bool] {
	return plan.Cast(config.Get("serverless"), plan.ToWeakBoolE)
}
