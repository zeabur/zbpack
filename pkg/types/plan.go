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
	PlanTypeStatic = "static"
)

type NodePackageManager string

const (
	NodePackageManagerYarn NodePackageManager = "yarn"
	NodePackageManagerPnpm NodePackageManager = "pnpm"
	NodePackageManagerNpm  NodePackageManager = "npm"
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
)

type PythonFramework string

const (
	PythonFrameworkFlask  PythonFramework = "flask"
	PythonFrameworkDjango PythonFramework = "django"
	PythonFrameworkNone   PythonFramework = "none"
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
	PhpFrameworkLaravel PhpFramework = "laravel"
	PhpFrameworkNone    PhpFramework = "none"
)
