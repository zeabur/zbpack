package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zeabur/zbpack/pkg/zeaburpack"
)

func main() {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	var absPath string
	if strings.HasPrefix(path, "https://") {
		absPath = path
	} else {
		absPath, _ = filepath.Abs(path)
	}

	submoduleName := filepath.Base(absPath)

	githubToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if strings.HasPrefix(absPath, "https://github.com") && githubToken == "" {
		panic("GITHUB_ACCESS_TOKEN is required for GitHub URL")
	}

	t, m := zeaburpack.Plan(
		zeaburpack.PlanOptions{
			SubmoduleName: &submoduleName,
			Path:          &absPath,
			AccessToken:   &githubToken,
		},
	)

	zeaburpack.PrintPlanAndMeta(t, m, func(log string) { println(log) })
}
