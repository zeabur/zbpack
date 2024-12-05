package transformer

import (
	"github.com/zeabur/zbpack/internal/nodejs/umi"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNodejsUmi transforms Node.js Umi.js functions.
func TransformNodejsUmi(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeNodejs || ctx.PlanMeta["framework"] != string(types.NodeProjectFrameworkUmi) || ctx.PlanMeta["serverless"] != "true" {
		return ErrSkip
	}

	return umi.TransformServerless(ctx.AppPath)
}
