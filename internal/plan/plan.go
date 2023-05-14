// Package plan is the interface for planners.
package plan

import (
	"github.com/zeabur/zbpack/internal/golang"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/dockerfile"
	"github.com/zeabur/zbpack/internal/rust"
	"github.com/zeabur/zbpack/pkg/types"

	"github.com/zeabur/zbpack/internal/deno"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/utils"
)

// Planner is the interface for planners.
type Planner interface {
	Plan() (types.PlanType, types.PlanMeta)
}

type planner struct {
	source             afero.Fs
	submoduleName      string
	customBuildCommand *string
	customStartCommand *string
	outputDir          *string
}

// NewPlannerOptions is the options for NewPlanner.
type NewPlannerOptions struct {
	Source             afero.Fs
	SubmoduleName      string
	CustomBuildCommand *string
	CustomStartCommand *string
	OutputDir          *string
}

// NewPlanner creates a new Planner.
func NewPlanner(opt *NewPlannerOptions) Planner {
	return &planner{
		source:             opt.Source,
		submoduleName:      opt.SubmoduleName,
		customBuildCommand: opt.CustomBuildCommand,
		customStartCommand: opt.CustomStartCommand,
		outputDir:          opt.OutputDir,
	}
}

func (b planner) Plan() (types.PlanType, types.PlanMeta) {
	// custom Dockerfile
	if utils.HasFile(b.source, "Dockerfile", "dockerfile") {
		return types.PlanTypeDocker, dockerfile.GetMeta(
			dockerfile.GetMetaOptions{
				Src: b.source,
			},
		)
	}

	// PHP project
	if utils.HasFile(b.source, "index.php", "composer.json") {
		framework := php.DetermineProjectFramework(b.source)
		phpVersion := php.GetPHPVersion(b.source)
		return types.PlanTypePHP, types.PlanMeta{
			"framework":  string(framework),
			"phpVersion": phpVersion,
		}
	}

	// Node.js project
	if utils.HasFile(b.source, "package.json") {
		return types.PlanTypeNodejs, nodejs.GetMeta(
			nodejs.GetMetaOptions{
				Src:            b.source,
				CustomBuildCmd: b.customBuildCommand,
				CustomStartCmd: b.customStartCommand,
				OutputDir:      b.outputDir,
			},
		)
	}

	// Go project
	if utils.HasFile(b.source, "go.mod") {
		return types.PlanTypeGo, golang.GetMeta(
			golang.GetMetaOptions{
				Src:           b.source,
				SubmoduleName: b.submoduleName,
			},
		)
	}

	// Python project
	if utils.HasFile(
		b.source,
		"app.py", "main.py", "app.py", "manage.py", "requirements.txt",
	) {
		return types.PlanTypePython, python.GetMeta(python.GetMetaOptions{Src: b.source})
	}

	// Ruby project
	if utils.HasFile(b.source, "Gemfile") {
		rubyVersion := ruby.DetermineRubyVersion(b.source)
		framework := ruby.DetermineRubyFramework(b.source)
		return types.PlanTypeRuby, types.PlanMeta{
			"rubyVersion": rubyVersion,
			"framework":   string(framework),
		}
	}

	// Java project
	if utils.HasFile(
		b.source, "pom.xml", "pom.yml", "pom.yaml", "build.gradle",
		"build.gradle.kts",
	) {
		projectType := java.DetermineProjectType(b.source)
		framework := java.DetermineFramework(projectType, b.source)
		jdkVersion := java.DetermineJDKVersion(projectType, b.source)
		return types.PlanTypeJava, types.PlanMeta{
			"type":      string(projectType),
			"framework": string(framework),
			"jdk":       jdkVersion,
		}
	}

	// Deno project
	if utils.HasFile(
		b.source, "deno.json", "deno.lock", "fresh.gen.ts",
	) {
		framework := deno.DetermineFramework(b.source)
		entry := deno.DetermineEntry(b.source)
		startCmd := deno.GetStartCommand(b.source)
		return types.PlanTypeDeno, types.PlanMeta{
			"framework":    string(framework),
			"entry":        entry,
			"startCommand": startCmd,
		}
	}

	// Rust project
	if utils.HasFile(b.source, "Cargo.toml") {
		return types.PlanTypeRust, rust.GetMeta(
			rust.GetMetaOptions{
				Src:           b.source,
				SubmoduleName: b.submoduleName,
			},
		)
	}

	// static site generator (hugo, gatsby, etc) detection
	if utils.HasFile(b.source, "index.html") {
		html, err := afero.ReadFile(b.source, "index.html")

		if err == nil && strings.Contains(string(html), "Hugo") {
			return types.PlanTypeStatic, types.PlanMeta{"framework": "hugo"}
		}

		if err == nil && strings.Contains(string(html), "Hexo") {
			return types.PlanTypeStatic, types.PlanMeta{"framework": "hexo"}
		}
	}

	return types.PlanTypeStatic, types.PlanMeta{}
}
