// Zbpack is a tool to help you build your project
// as Docker image in one click.
package main

import (
	"os"
	"path/filepath"

	"github.com/zeabur/zbpack/pkg/types"

	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

func main() {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	trueValue := true

	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	submoduleName := filepath.Base(absPath)

	handlePlanDetermined := func(planType types.PlanType, planMeta types.PlanMeta) {
		zeaburpack.PrintPlanAndMeta(
			planType, planMeta, func(log string) {
				println(log)
			},
		)
	}

	err = zeaburpack.Build(
		&zeaburpack.BuildOptions{
			Path:                 &path,
			Interactive:          &trueValue,
			HandlePlanDetermined: &handlePlanDetermined,
			SubmoduleName:        &submoduleName,
		},
	)

	if err != nil {
		panic(err)
	}
}
