// Package zbserverless provides the serverless transformer for Zeabur.
package zbserverless

import "github.com/zeabur/zbpack/pkg/types"

// ServerlessTransformer turns the build output into Zeabur serverless format.
//
// The imageRootDirectory is the directory of the built image.
// The directory structure is the same as what your Dockerfile specifies.
//
// The dotZeaburDirectory is the directory to write the serverless artifact to.
// Usually, it is the ".zeabur" directory in the project root.
//
// The planMeta is the metadata of the build plan.
//
// The function should return an error if the transformation fails.
type ServerlessTransformer func(imageRootDirectory string, dotZeaburDirectory string, planMeta types.PlanMeta) error
