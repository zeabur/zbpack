// Zbpack is a tool to help you build your project
// as Docker image in one click.
package main

import (
	"log"

	"github.com/salamer/zbpack/internal/zbpack"
)

func main() {
	if err := zbpack.Execute(); err != nil {
		log.Fatalln(err)
	}
}
