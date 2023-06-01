// Zbpack is a tool to help you build your project
// as Docker image in one click.
package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/zeabur/zbpack/pkg/types"

	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

func main() {
	validate()

	path := os.Args[1]

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

// validate function performs some necessary checks before starting.
func validate() {
	if len(os.Args) < 2 {
		println("Usage: zbpack <directory to analyse or build>")
		os.Exit(0)
	}

	c, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		println("Failed to create Docker client: " + err.Error())
		os.Exit(0)
	}

	_, err = c.Ping(context.Background())
	if err != nil {
		println("Important: Please make sure the Docker daemon is running.")
		os.Exit(0)
	}

	defer func(c *client.Client) {
		err := c.Close()
		if err != nil {
			println("Failed to close Docker client: " + err.Error())
		}
	}(c)
}
