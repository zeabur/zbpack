package main

import (
	"os"
	"path/filepath"

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

	err = zeaburpack.Build(
		&zeaburpack.BuildOptions{
			Path:        &path,
			Interactive: &trueValue,

			SubmoduleName: &submoduleName,
		},
	)
	if err != nil {
		panic(err)
	}
}
