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

	err := zeaburpack.Build(
		&zeaburpack.BuildOptions{
			Path: &path,
		},
	)

	if err != nil {
		panic(err)
	}
}
