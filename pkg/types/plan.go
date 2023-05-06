package types

// PlanType is primary category of the build plan.
// For example, the programing language or the runtime.
type PlanType string

// PlanMeta is the metadata for the build plan.
// For example, the runtime version, the package manager or the framework used.
type PlanMeta map[string]string

const (
	PlanTypeNodejs = "nodejs"
	PlanTypeGo     = "go"
	PlanTypePython = "python"
	PlanTypeRuby   = "ruby"
	PlanTypeDocker = "docker"
	PlanTypePhp    = "php"
	PlanTypeJava   = "java"
	PlanTypeDeno   = "deno"
	PlanTypeRust   = "rust"
	PlanTypeStatic = "static"
)

type NodePackageManager string

const (
	NodePackageManagerYarn    NodePackageManager = "yarn"
	NodePackageManagerPnpm    NodePackageManager = "pnpm"
	NodePackageManagerNpm     NodePackageManager = "npm"
	NodePackageManagerUnknown NodePackageManager = "unknown"
)

type NodeProjectFramework string

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
)

type PythonFramework string

const (
	PythonFrameworkFlask   PythonFramework = "flask"
	PythonFrameworkDjango  PythonFramework = "django"
	PythonFrameworkFastapi PythonFramework = "fastapi"
	PythonFrameworkNone    PythonFramework = "none"
)

type JavaProjectType string

const (
	JavaProjectTypeMaven  JavaProjectType = "maven"
	JavaProjectTypeGradle JavaProjectType = "gradle"
	JavaProjectTypeNone   JavaProjectType = "none"
)

type JavaFramework string

const (
	JavaFrameworkSpringBoot JavaFramework = "spring-boot"
	JavaFrameworkNone       JavaFramework = "none"
)

type PhpFramework string

const (
	PhpFrameworkLaravel     PhpFramework = "laravel"
	PhpFrameworkNone        PhpFramework = "none"
	PhpFrameworkThinkphp    PhpFramework = "thinkphp"
	PhpFrameworkCodeigniter PhpFramework = "codeigniter"
)

type RubyFramework string

const (
	RubyFrameworkRails RubyFramework = "rails"
	RubyFrameworkNone  RubyFramework = "none"
)

type DenoFramework string

const (
	DenoFrameworkFresh DenoFramework = "fresh"
	DenoFrameworkNone  DenoFramework = "none"
)
