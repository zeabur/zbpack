package plan

import (
	"os"
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
	absPath            string
	submoduleName      string
	customBuildCommand *string
	customStartCommand *string
	outputDir          *string
}

type NewPlannerOptions struct {
	AbsPath            string
	SubmoduleName      string
	CustomBuildCommand *string
	CustomStartCommand *string
	OutputDir          *string
}

func NewPlanner(opt *NewPlannerOptions) Planner {
	return &planner{
		absPath:            opt.AbsPath,
		submoduleName:      opt.SubmoduleName,
		customBuildCommand: opt.CustomBuildCommand,
		customStartCommand: opt.CustomStartCommand,
		outputDir:          opt.OutputDir,
	}
}

func (b planner) Plan() (PlanType, PlanMeta) {
	// custom Dockerfile
	if utils.HasFile(b.absPath, "Dockerfile", "dockerfile") {
		return PlanTypeDocker, dockerfile.GetMeta(
			dockerfile.GetMetaOptions{
				AbsPath: b.absPath,
			},
		)
	}

	// PHP project
	if utils.HasFile(b.absPath, "index.php", "composer.json") {
		framework := php.DetermineProjectFramework(b.absPath)
		phpVersion := php.GetPhpVersion(b.absPath)
		return PlanTypePhp, PlanMeta{
			"framework":  string(framework),
			"phpVersion": phpVersion,
		}
	}

	// Node.js project
	if utils.HasFile(b.absPath, "package.json") {
		return PlanTypeNodejs, nodejs.GetMeta(
			nodejs.GetMetaOptions{
				AbsPath:        b.absPath,
				CustomBuildCmd: b.customBuildCommand,
				CustomStartCmd: b.customStartCommand,
				OutputDir:      b.outputDir,
			},
		)
	}

	// Go project
	if utils.HasFile(b.absPath, "go.mod") {

		// in a basic go project, we assume the entrypoint is main.go in root directory
		if utils.HasFile(b.absPath, "main.go") {
			return PlanTypeGo, PlanMeta{"entry": "main.go"}
		}

		// if there is no main.go in root directory, we assume it's a monorepo project.
		// in a general monorepo Go repo of service "user-service", the entry point might be `./cmd/user-service/main.go`
		if utils.HasFile(
			path.Join(b.absPath, "cmd", b.submoduleName), "main.go",
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
		b.absPath,
		"app.py", "main.py", "app.py", "manage.py", "requirements.txt",
	) {
		framework := python.DetermineFramework(b.absPath)
		entry := python.DetermineEntry(b.absPath)
		dependencyPolicy := python.DetermineDependencyPolicy(b.absPath)
		return PlanTypePython, PlanMeta{
			"framework":        string(framework),
			"entry":            entry,
			"dependencyPolicy": dependencyPolicy,
		}
	}

	// Ruby project
	if utils.HasFile(b.absPath, "Gemfile") {
		rubyVersion := ruby.DetermineRubyVersion(b.absPath)
		framework := ruby.DetermineRubyFramework(b.absPath)
		return PlanTypeRuby, PlanMeta{
			"rubyVersion": rubyVersion,
			"framework":   string(framework),
		}
	}

	// Java project
	if utils.HasFile(
		b.absPath, "pom.xml", "pom.yml", "pom.yaml", "build.gradle",
		"build.gradle.kts",
	) {
		projectType := java.DetermineProjectType(b.absPath)
		framework := java.DetermineFramework(projectType, b.absPath)
		jdkVersion := java.DetermineJDKVersion(projectType, b.absPath)
		return PlanTypeJava, PlanMeta{
			"type":      string(projectType),
			"framework": string(framework),
			"jdk":       jdkVersion,
		}
	}

	// Deno project
	if utils.HasFile(
		b.absPath, "deno.json", "deno.lock", "fresh.gen.ts",
	) {
		framework := deno.DetermineFramework(b.absPath)
		entry := deno.DetermineEntry(b.absPath)
		startCmd := deno.GetStartCommand(b.absPath)
		return PlanTypeDeno, PlanMeta{
			"framework":    string(framework),
			"entry":        entry,
			"startCommand": startCmd,
		}
	}

	// Rust project
	if utils.HasFile(b.absPath, "Cargo.toml") {
		return PlanTypeRust, rust.GetMeta(rust.GetMetaOptions{
			AbsPath:       b.absPath,
			SubmoduleName: b.submoduleName,
		})
	}

	// static site generator (hugo, gatsby, etc) detection
	if utils.HasFile(b.absPath, "index.html") {
		htmlPath := path.Join(b.absPath, "index.html")
		html, err := os.ReadFile(htmlPath)

		if err == nil && strings.Contains(string(html), "Hugo") {
			return PlanTypeStatic, PlanMeta{"framework": "hugo"}
		}

		if err == nil && strings.Contains(string(html), "Hexo") {
			return PlanTypeStatic, PlanMeta{"framework": "hexo"}
		}
	}

	return PlanTypeStatic, PlanMeta{}
}
