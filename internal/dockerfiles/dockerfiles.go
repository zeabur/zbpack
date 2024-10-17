// Package dockerfiles provides Dockerfile templates for different languages and frameworks.
package dockerfiles

import (
	"embed"
	_ "embed"
	"path"
)

//go:embed composable
var composable embed.FS

// GetDockerfileContent returns the content of the Dockerfile with the given name.
func GetDockerfileContent(name string) ([]byte, error) {
	return composable.ReadFile(path.Join("composable", name+".Dockerfile"))
}
