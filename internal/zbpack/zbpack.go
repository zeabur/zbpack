// Package zbpack is internal package, contain the main logic of zbpack command-line interface.
package zbpack

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

var (
	// info option is used to analyze and print project information.
	info bool
	// dockerfile option is used to generate a Dockerfile.
	dockerfile bool
	// userSubmoduleName option is used to specify the submodule name of this project manually
	userSubmoduleName string
	cmd               = &cobra.Command{
		Use:   "zbpack",
		Short: "Zbpack is a tool to help you analyze your project and build Docker image in one click.",
		Long: "Zbpack is a powerful tool that not only analyzes your project for dependencies and requirements, " +
			"but also builds Docker images in one click, greatly simplifying your workflow.",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return fmt.Errorf("zbpack requires a directory to analyse or build")
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args)
		},
	}
)

func init() {
	cmd.PersistentFlags().BoolVarP(&info, "info", "i", false, "only print project information")
	cmd.PersistentFlags().BoolVarP(&dockerfile, "dockerfile", "d", false, "output dockerfile")
	cmd.PersistentFlags().StringVar(&userSubmoduleName, "submodule", "", "submodule (service) name. by default, it is picked from the directory name.")
	cmd.SetUsageTemplate(usageTemplate)
}

// Execute is used to execute zbpack command-line interface.
func Execute() error {
	return cmd.Execute()
}

// run is command-line entrypoint.
func run(args []string) error {
	path := args[0]

	switch {
	case info:
		return plan(path)
	case dockerfile:
		return PlanAndOutputDockerfile(path)
	default:
		return build(path)
	}
}

// build is used to build Docker image and show build plan.
func build(path string) error {
	// before start, check if buildctl is installed and buildkitd is running
	err := exec.Command("buildctl", "debug", "workers").Run()
	if err != nil {
		red := "\033[31m"
		blue := "\033[34m"
		reset := "\033[0m"
		gray := "\033[90m"

		print(red, "buildctl is not installed or buildkitd is not running.\n", reset)
		print("Learn more: https://github.com/moby/buildkit#quick-start\n\n", reset)
		print(gray, "Or you can simply run the following command to run buildkitd in a container:\n", reset)
		print(blue, "docker run -d --name buildkitd --privileged moby/buildkit:latest\n\n", reset)
		print(gray, "And then install buildctl if you haven't:\n", reset)
		print(blue, "docker cp buildkitd:/usr/bin/buildctl /usr/local/bin\n\n", reset)
		print(gray, "After that, you can run zbpack again with the following command:\n", reset)
		print(blue, "BUILDKIT_HOST=docker-container://buildkitd zbpack <...>\n", reset)

		return nil
	}

	// TODO support online repositories
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return fmt.Errorf("zbpack does not support building from online repositories yet")
	}

	submoduleName, err := GetSubmoduleName(path)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("using submoduleName: %s", submoduleName)

	userVarsList := os.Environ()
	userVarsToBuild := make(map[string]string)
	for _, userVar := range userVarsList {
		key, value, ok := strings.Cut(userVar, "=")
		if !ok {
			continue
		}

		if key, ok := strings.CutPrefix(key, "ZBPACK_VAR_"); ok {
			userVarsToBuild[key] = value
		}
	}

	log.Printf("environment variables to pass: %+v", userVarsToBuild)

	return zeaburpack.Build(
		&zeaburpack.BuildOptions{
			Path:          &path,
			Interactive:   lo.ToPtr(true),
			SubmoduleName: &submoduleName,
			UserVars:      &userVarsToBuild,
		},
	)
}

// plan is used to analyze and print project information.
func plan(path string) error {
	submoduleName, err := GetSubmoduleName(path)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("using submoduleName: %s", submoduleName)

	var githubToken *string
	githubTokenStr := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubTokenStr != "" {
		githubToken = &githubTokenStr
	}

	t, m := zeaburpack.Plan(
		zeaburpack.PlanOptions{
			SubmoduleName: &submoduleName,
			Path:          &path,
			AccessToken:   githubToken,
		},
	)

	zeaburpack.PrintPlanAndMeta(t, m, os.Stderr)

	return nil
}

// PlanAndOutputDockerfile is used to generate Dockerfile and output it.
func PlanAndOutputDockerfile(path string) error {
	submoduleName, err := GetSubmoduleName(path)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("using submoduleName: %s", submoduleName)

	var githubToken *string
	githubTokenStr := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubTokenStr != "" {
		githubToken = &githubTokenStr
	}

	// Plan and output Dockerfile
	return zeaburpack.PlanAndOutputDockerfile(
		zeaburpack.PlanOptions{
			SubmoduleName: &submoduleName,
			Path:          &path,
			AccessToken:   githubToken,
		},
	)
}
