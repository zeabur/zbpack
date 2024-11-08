// Package php is the planner for PHP projects.
package php

import (
	"fmt"
	"strings"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"

	_ "embed"
)

//go:embed Dockerfile
var dockerfile string

// GenerateDockerfile generates the Dockerfile for PHP projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	compiledDockerfile := dockerfile

	variables := map[string]string{
		"PHP_VERSION":            meta["phpVersion"],
		"APT_EXTRA_DEPENDENCIES": meta["deps"],
		"PHP_EXTENSIONS":         meta["exts"],
		"BUILD_COMMAND":          meta["buildCommand"],
		"START_COMMAND":          meta["startCommand"],
		"PHP_OPTIMIZE":           meta["optimize"],
	}

	for k, v := range variables {
		compiledDockerfile = strings.Replace(compiledDockerfile, "ARG "+k, fmt.Sprintf("ARG %s=%q", k, v), 1)
	}

	return compiledDockerfile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new PHP packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
