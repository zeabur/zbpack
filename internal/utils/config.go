package utils

import (
	"os"

	"github.com/moznion/go-optional"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/pkg/plan"
)

// GetExplicitServerlessConfig gets the serverless flag from the project configuration.
//
// When the serverless flag is not set, it will be determined by the FORCE_CONTAINERIZED
// and ZBPACK_SERVERLESS environment variables.
// If all of them are not set, it returns None for consumers to determine the default value.
func GetExplicitServerlessConfig(config plan.ImmutableProjectConfiguration) optional.Option[bool] {
	serverlessConfig := plan.Cast(config.Get("serverless"), cast.ToBoolE)
	if value, err := serverlessConfig.Take(); err == nil {
		return optional.Some(value)
	}

	fcEnv := os.Getenv("FORCE_CONTAINERIZED")
	if fcEnv == "true" || fcEnv == "1" {
		return optional.Some(true)
	}

	zsEnv := os.Getenv("ZBPACK_SERVERLESS")
	if zsEnv == "true" || zsEnv == "1" {
		return optional.Some(true)
	}

	return optional.None[bool]()
}
