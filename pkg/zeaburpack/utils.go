package zeaburpack

import (
	"fmt"
	"github.com/zeabur/zbpack/internal/types"
)

func PrintPlanAndMeta(plan types.PlanType, meta types.PlanMeta, handleLog func(log string)) {
	handleLog("========== Plan determined ==========")
	handleLog("Plan type: " + string(plan))
	handleLog("Plan meta: ")
	for k, v := range meta {
		handleLog(fmt.Sprintf("  %s: \"%s\"", k, v))
	}
	handleLog("\n")
}
