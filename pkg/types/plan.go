// Package types is the type definitions for the build plan in Zbpack.
package types

import "slices"

// PlanType is primary category of the build plan.
// For example, the programing language or the runtime.
type PlanType string

// PlanMeta is the metadata for the build plan.
// For example, the runtime version, the package manager or the framework used.
type PlanMeta map[string]string

//revive:disable:exported
const (
	PlanTypeNodejs PlanType = "nodejs"
	PlanTypeGo     PlanType = "go"
	PlanTypePython PlanType = "python"
	PlanTypeRuby   PlanType = "ruby"
	PlanTypeDocker PlanType = "docker"
	PlanTypePHP    PlanType = "php"
	PlanTypeJava   PlanType = "java"
	PlanTypeDeno   PlanType = "deno"
	PlanTypeRust   PlanType = "rust"
	PlanTypeDotnet PlanType = "dotnet"
	PlanTypeElixir PlanType = "elixir"
	PlanTypeGleam  PlanType = "gleam"
	PlanTypeBun    PlanType = "bun"
	PlanTypeStatic PlanType = "static"
	PlanTypeSwift  PlanType = "swift"
	PlanTypeDart   PlanType = "dart"
	PlanTypeNix    PlanType = "nix"
)

type DartFramework string

const (
	DartFrameworkFlutter   DartFramework = "flutter"
	DartFrameworkServerpod DartFramework = "serverpod"
	DartFrameworkNone      DartFramework = "none"
)

//revive:enable:exported

// NodePackageManager represents the package manager used in a Node.js project.
type NodePackageManager string

//revive:disable:exported
const (
	NodePackageManagerYarn    NodePackageManager = "yarn"
	NodePackageManagerPnpm    NodePackageManager = "pnpm"
	NodePackageManagerNpm     NodePackageManager = "npm"
	NodePackageManagerBun     NodePackageManager = "bun"
	NodePackageManagerUnknown NodePackageManager = "unknown"
)

//revive:enable:exported

// NodeProjectFramework represents the framework of a Node.js project.
type NodeProjectFramework string

//revive:disable:exported
const (
	NodeProjectFrameworkWaku             NodeProjectFramework = "waku"
	NodeProjectFrameworkVite             NodeProjectFramework = "vite"
	NodeProjectFrameworkQwik             NodeProjectFramework = "qwik"
	NodeProjectFrameworkCreateReactApp   NodeProjectFramework = "create-react-app"
	NodeProjectFrameworkNextJs           NodeProjectFramework = "next.js"
	NodeProjectFrameworkRemix            NodeProjectFramework = "remix"
	NodeProjectFrameworkNuxtJs           NodeProjectFramework = "nuxt.js"
	NodeProjectFrameworkVueCli           NodeProjectFramework = "vue-cli"
	NodeProjectFrameworkAngular          NodeProjectFramework = "angular"
	NodeProjectFrameworkNone             NodeProjectFramework = "none"
	NodeProjectFrameworkUmi              NodeProjectFramework = "umi"
	NodeProjectFrameworkSvelte           NodeProjectFramework = "svelte"
	NodeProjectFrameworkNestJs           NodeProjectFramework = "nest.js"
	NodeProjectFrameworkHexo             NodeProjectFramework = "hexo"
	NodeProjectFrameworkVitepress        NodeProjectFramework = "vitepress"
	NodeProjectFrameworkAstro            NodeProjectFramework = "astro"
	NodeProjectFrameworkAstroStatic      NodeProjectFramework = "astro-static"
	NodeProjectFrameworkAstroSSR         NodeProjectFramework = "astro-ssr"
	NodeProjectFrameworkAstroStarlight   NodeProjectFramework = "astro-starlight"
	NodeProjectFrameworkSliDev           NodeProjectFramework = "sli.dev"
	NodeProjectFrameworkDocusaurus       NodeProjectFramework = "docusaurus"
	NodeProjectFrameworkSolidStart       NodeProjectFramework = "solid-start"
	NodeProjectFrameworkSolidStartVinxi  NodeProjectFramework = "solid-start-vinxi"
	NodeProjectFrameworkSolidStartNode   NodeProjectFramework = "solid-start-node"
	NodeProjectFrameworkSolidStartStatic NodeProjectFramework = "solid-start-static"
	NodeProjectFrameworkNueJs            NodeProjectFramework = "nuejs"
	NodeProjectFrameworkVocs             NodeProjectFramework = "vocs"
	NodeProjectFrameworkRspress          NodeProjectFramework = "rspress"
	NodeProjectFrameworkGrammY           NodeProjectFramework = "grammy"
	NodeProjectFrameworkNitropack        NodeProjectFramework = "nitropack"
)

var NitroBasedFrameworks = []NodeProjectFramework{
	NodeProjectFrameworkNuxtJs,
	NodeProjectFrameworkNitropack,
	NodeProjectFrameworkSolidStartVinxi,
}

func IsNitroBasedFramework(framework string) bool {
	return slices.ContainsFunc(NitroBasedFrameworks, func(f NodeProjectFramework) bool {
		return string(f) == framework
	})
}

//revive:enable:exported

// PythonFramework represents the framework of a Python project.
type PythonFramework string

//revive:disable:exported
const (
	PythonFrameworkFlask   PythonFramework = "flask"
	PythonFrameworkDjango  PythonFramework = "django"
	PythonFrameworkFastapi PythonFramework = "fastapi"
	PythonFrameworkTornado PythonFramework = "tornado"
	PythonFrameworkSanic   PythonFramework = "sanic"
	PythonFrameworkNone    PythonFramework = "none"

	// PythonFrameworkStreamlit https://github.com/streamlit/streamlit
	PythonFrameworkStreamlit PythonFramework = "streamlit"

	// PythonFrameworkReflex https://github.com/reflex-dev/reflex
	PythonFrameworkReflex PythonFramework = "reflex"
)

//revive:enable:exported

// PythonPackageManager is the type of the package manager.
type PythonPackageManager string

//revive:disable:exported
const (
	PythonPackageManagerUnknown PythonPackageManager = "unknown"
	PythonPackageManagerPip     PythonPackageManager = "pip"
	PythonPackageManagerPoetry  PythonPackageManager = "poetry"
	PythonPackageManagerPipenv  PythonPackageManager = "pipenv"
	PythonPackageManagerPdm     PythonPackageManager = "pdm"
	PythonPackageManagerRye     PythonPackageManager = "rye"
	PythonPackageManagerUv      PythonPackageManager = "uv"
)

type SwiftFramework string

//revive:disable:exported
const (
	SwiftFrameworkVapor SwiftFramework = "vapor"
	SwiftFrameworkNone  SwiftFramework = "none"
)

//revive:enable:exported

// JavaProjectType represents the type of a Java project.
type JavaProjectType string

//revive:disable:exported
const (
	JavaProjectTypeMaven  JavaProjectType = "maven"
	JavaProjectTypeGradle JavaProjectType = "gradle"
	JavaProjectTypeNone   JavaProjectType = "none"
)

//revive:enable:exported

// JavaFramework represents the framework of a Java project.
type JavaFramework string

//revive:disable:exported
const (
	JavaFrameworkSpringBoot JavaFramework = "spring-boot"
	JavaFrameworkNone       JavaFramework = "none"
)

//revive:enable:exported

// PHPFramework represents the framework of a PHP project.
type PHPFramework string

//revive:disable:exported
const (
	PHPFrameworkLaravel     PHPFramework = "laravel"
	PHPFrameworkNone        PHPFramework = "none"
	PHPFrameworkThinkphp    PHPFramework = "thinkphp"
	PHPFrameworkCodeigniter PHPFramework = "codeigniter"
	PHPFrameworkSymfony     PHPFramework = "symfony"
)

//revive:enable:exported

// RubyFramework represents the framework of a Ruby project.
//
//revive:enable:exported
type RubyFramework string

//revive:disable:exported
const (
	RubyFrameworkRails RubyFramework = "rails"
	RubyFrameworkNone  RubyFramework = "none"
)

//revive:enable:exported

// DenoFramework represents the framework of a Deno project.
type DenoFramework string

//revive:disable:exported
const (
	DenoFrameworkFresh DenoFramework = "fresh"
	DenoFrameworkNone  DenoFramework = "none"
)

//revive:enable:exported

// DotnetFramework represents the framework of a Dotnet project.
type DotnetFramework string

//revive:disable:exported
const (
	DotnetFrameworkAspnet     DotnetFramework = "aspnet"
	DotnetFrameworkBlazorWasm DotnetFramework = "blazorwasm"
	DotnetFrameworkConsole    DotnetFramework = "console"
)

//revive:enable:exported

// ElixirFramework represents the framework of a Elixir project.
type ElixirFramework string

//revive:disable:exported
const (
	ElixirFrameworkPhoenix ElixirFramework = "phoenix"
)

//revive:enable:exported

// BunFramework represents the framework of a Bun project.
type BunFramework string

//revive:enable:exported
const (
	BunFrameworkElysia BunFramework = "elysia"
	BunFrameworkBaojs  BunFramework = "baojs"
	BunFrameworkBagel  BunFramework = "bagel"
	BunFrameworkHono   BunFramework = "hono"
	BunFrameworkNone   BunFramework = "none"
)
