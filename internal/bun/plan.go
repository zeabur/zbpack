package bun

import (
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/pkg/types"
)

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions nodejs.GetMetaOptions

// GetMeta gets the metadata of the Node.js project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	return nodejs.GetMeta(nodejs.GetMetaOptions(opt))
}
