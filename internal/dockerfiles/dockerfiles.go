// Package dockerfiles provides Dockerfile templates for different languages and frameworks.
package dockerfiles

import (
	"embed"
	_ "embed"
	"path"
)

//go:embed selectable
var selectable embed.FS

// GetDockerfileContent returns the content of the Dockerfile with the given name.
func GetDockerfileContent(name string) ([]byte, error) {
	return selectable.ReadFile(path.Join("selectable", name+".Dockerfile"))
}
