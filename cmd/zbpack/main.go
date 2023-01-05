package main

import (
	"github.com/zeabur/zbpack/pkg/zeaburpack"
	"os"
)

func main() {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	trueValue := true

	err := zeaburpack.Build(
		&zeaburpack.BuildOptions{
			Path:        &path,
			Interactive: &trueValue,
		},
	)

	if err != nil {
		panic(err)
	}
}
