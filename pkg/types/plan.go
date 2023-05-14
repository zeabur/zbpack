// Package types is the type definitions for the build plan in Zbpack.
package types

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
	PlanTypeStatic PlanType = "static"
)

//revive:enable:exported

// NodePackageManager represents the package manager used in a Node.js project.
type NodePackageManager string

//revive:disable:exported
const (
	NodePackageManagerYarn    NodePackageManager = "yarn"
	NodePackageManagerPnpm    NodePackageManager = "pnpm"
	NodePackageManagerNpm     NodePackageManager = "npm"
	NodePackageManagerUnknown NodePackageManager = "unknown"
)

//revive:enable:exported

// NodeProjectFramework represents the framework of a Node.js project.
type NodeProjectFramework string

//revive:disable:exported
const (
	NodeProjectFrameworkVite           NodeProjectFramework = "vite"
	NodeProjectFrameworkQwik           NodeProjectFramework = "qwik"
	NodeProjectFrameworkCreateReactApp NodeProjectFramework = "create-react-app"
	NodeProjectFrameworkNextJs         NodeProjectFramework = "next.js"
	NodeProjectFrameworkRemix          NodeProjectFramework = "remix"
	NodeProjectFrameworkNuxtJs         NodeProjectFramework = "nuxt.js"
	NodeProjectFrameworkVueCli         NodeProjectFramework = "vue-cli"
	NodeProjectFrameworkNone           NodeProjectFramework = "none"
	NodeProjectFrameworkUmi            NodeProjectFramework = "umi"
	NodeProjectFrameworkSvelte         NodeProjectFramework = "svelte"
	NodeProjectFrameworkNestJs         NodeProjectFramework = "nest.js"
	NodeProjectFrameworkHexo           NodeProjectFramework = "hexo"
	NodeProjectFrameworkVitepress      NodeProjectFramework = "vitepress"
	NodeProjectFrameworkAstroStatic    NodeProjectFramework = "astro-static"
	NodeProjectFrameworkAstroSSR       NodeProjectFramework = "astro-ssr"
	NodeProjectFrameworkSliDev         NodeProjectFramework = "sli.dev"
)

//revive:enable:exported

// PythonFramework represents the framework of a Python project.
type PythonFramework string

//revive:disable:exported
const (
	PythonFrameworkFlask   PythonFramework = "flask"
	PythonFrameworkDjango  PythonFramework = "django"
	PythonFrameworkFastapi PythonFramework = "fastapi"
	PythonFrameworkNone    PythonFramework = "none"
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
)

//revive:enable:exported

// RubyFramework represents the framework of a Ruby project.
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
