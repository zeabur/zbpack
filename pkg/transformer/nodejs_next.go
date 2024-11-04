package transformer

import (
	"github.com/zeabur/zbpack/internal/nodejs/nextjs"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNodejsNext transforms Node.js Next.js functions.
func TransformNodejsNext(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeNodejs || ctx.PlanMeta["framework"] != string(types.NodeProjectFrameworkNextJs) || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	// fixme: too complex to migrate to pkg/transformer
	return nextjs.TransformServerless(ctx.AppPath)
}
