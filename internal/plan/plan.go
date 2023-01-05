package plan

import (
	"os"
	"path"
	"strings"

	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

type Planner interface {
	Plan() (PlanType, PlanMeta)
}

type planner struct {
	absPath       string
	submoduleName string
}

func NewPlanner(absPath string, submoduleName string) Planner {
	return &planner{absPath, submoduleName}
}

func (b planner) Plan() (PlanType, PlanMeta) {

	// custom Dockerfile
	if utils.HasFile(b.absPath, "Dockerfile", "dockerfile") {
		return PlanTypeDocker, PlanMeta{}
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
		pkgManager := nodejs.DeterminePackageManager(b.absPath)
		framework := nodejs.DetermineProjectFramework(b.absPath)
		buildCmd := nodejs.GetBuildCommand(b.absPath)
		startCmd := nodejs.GetStartCommand(b.absPath)
		mainFile := nodejs.GetMainFile(b.absPath)
		nodeVersion := nodejs.GetNodeVersion(b.absPath)
		needPuppeteer := nodejs.DetermineNeedPuppeteer(b.absPath)
		return PlanTypeNodejs, PlanMeta{
			"packageManager": string(pkgManager),
			"framework":      string(framework),
			"buildCommand":   buildCmd,
			"startCommand":   startCmd,
			"mainFile":       mainFile,
			"nodeVersion":    nodeVersion,
			"needPuppeteer":  needPuppeteer,
		}
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
		return PlanTypeRuby, PlanMeta{}
	}

	// Java project
	if utils.HasFile(
		b.absPath, "pom.xml", "pom.yml", "pom.yaml", "build.gradle",
		"build.gradle.kts",
	) {
		projectType := java.DetermineProjectType(b.absPath)
		framework := java.DetermineFramework(projectType, b.absPath)
		return PlanTypeJava, PlanMeta{
			"type":      string(projectType),
			"framework": string(framework),
		}
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
