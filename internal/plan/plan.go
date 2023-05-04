package plan

import (
	"github.com/zeabur/zbpack/internal/source"
	"path"
	"strings"

	"github.com/zeabur/zbpack/internal/dockerfile"
	"github.com/zeabur/zbpack/internal/rust"

	"github.com/zeabur/zbpack/internal/deno"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

type Planner interface {
	Plan() (PlanType, PlanMeta)
}

type planner struct {
	source             *source.Source
	submoduleName      string
	customBuildCommand *string
	customStartCommand *string
	outputDir          *string
}

type NewPlannerOptions struct {
	Source             *source.Source
	SubmoduleName      string
	CustomBuildCommand *string
	CustomStartCommand *string
	OutputDir          *string
}

func NewPlanner(opt *NewPlannerOptions) Planner {
	return &planner{
		source:             opt.Source,
		submoduleName:      opt.SubmoduleName,
		customBuildCommand: opt.CustomBuildCommand,
		customStartCommand: opt.CustomStartCommand,
		outputDir:          opt.OutputDir,
	}
}

func (b planner) Plan() (PlanType, PlanMeta) {
	// custom Dockerfile
	if utils.HasFile(b.source, "Dockerfile", "dockerfile") {
		return PlanTypeDocker, dockerfile.GetMeta(
			dockerfile.GetMetaOptions{
				Src: b.source,
			},
		)
	}

	// PHP project
	if utils.HasFile(b.source, "index.php", "composer.json") {
		framework := php.DetermineProjectFramework(b.source)
		phpVersion := php.GetPhpVersion(b.source)
		return PlanTypePhp, PlanMeta{
			"framework":  string(framework),
			"phpVersion": phpVersion,
		}
	}

	// Node.js project
	if utils.HasFile(b.source, "package.json") {
		return PlanTypeNodejs, nodejs.GetMeta(
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

		// in a basic go project, we assume the entrypoint is main.go in root directory
		if utils.HasFile(b.source, "main.go") {
			return PlanTypeGo, PlanMeta{"entry": "main.go"}
		}

		// if there is no main.go in root directory, we assume it's a monorepo project.
		// in a general monorepo Go repo of service "user-service", the entry point might be `./cmd/user-service/main.go`
		if utils.HasFile(
			b.source, path.Join("cmd", b.submoduleName, "main.go"),
		) {
			entry := path.Join("cmd", b.submoduleName, "main.go")
			return PlanTypeGo, PlanMeta{"entry": entry}
		}

		// We know it's a Go project, but we don't know how to build it.
		// We'll just return a generic Go plan type.
		return PlanTypeGo, PlanMeta{}
	}

	// Python project
	if utils.HasFile(
		b.source,
		"app.py", "main.py", "app.py", "manage.py", "requirements.txt",
	) {
		return PlanTypePython, python.GetMeta(python.GetMetaOptions{Src: b.source})
	}

	// Ruby project
	if utils.HasFile(b.source, "Gemfile") {
		rubyVersion := ruby.DetermineRubyVersion(b.source)
		framework := ruby.DetermineRubyFramework(b.source)
		return PlanTypeRuby, PlanMeta{
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
		return PlanTypeJava, PlanMeta{
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
		return PlanTypeDeno, PlanMeta{
			"framework":    string(framework),
			"entry":        entry,
			"startCommand": startCmd,
		}
	}

	// Rust project
	if utils.HasFile(b.source, "Cargo.toml") {
		return PlanTypeRust, rust.GetMeta(
			rust.GetMetaOptions{
				Src:           b.source,
				SubmoduleName: b.submoduleName,
			},
		)
	}

	// static site generator (hugo, gatsby, etc) detection
	if utils.HasFile(b.source, "index.html") {
		html, err := (*b.source).ReadFile("index.html")

		if err == nil && strings.Contains(string(html), "Hugo") {
			return PlanTypeStatic, PlanMeta{"framework": "hugo"}
		}

		if err == nil && strings.Contains(string(html), "Hexo") {
			return PlanTypeStatic, PlanMeta{"framework": "hexo"}
		}
	}

	return PlanTypeStatic, PlanMeta{}
}
