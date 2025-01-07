package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/zeabur/zbpack/pkg/types"
	"github.com/zeabur/zbpack/pkg/zeaburpack"
	"golang.org/x/exp/maps"
)

var projects = []struct {
	name  string
	owner string
	repo  string
	dir   string
}{
	/* bun */
	{
		name:  "bun-hono",
		owner: "zeabur",
		repo:  "hono-bun-template",
	},
	{
		name:  "bun-nextjs",
		owner: "zeabur",
		repo:  "nextjs-bun-template",
	},
	{
		name:  "bun-nuxtjs",
		owner: "zeabur",
		repo:  "nuxtjs-bun-template",
	},
	{
		name:  "bun-baojs",
		owner: "zeabur",
		repo:  "baojs-template",
	},
	{
		name:  "bun-elysia",
		owner: "zeabur",
		repo:  "elysia-starter",
	},
	{
		name: "bun-plain",
		dir:  "bun-plain",
	},
	{
		name: "bun-without-lockfile",
		dir:  "bun-without-lockfile",
	},
	{
		name: "bun-yarn-lockfile",
		dir:  "bun-yarn-lockfile",
	},
	{
		name:  "bun-bagel",
		owner: "zeabur",
		repo:  "bagel-template",
	},

	/* dart */
	{
		name:  "dart-flutter",
		owner: "zeabur",
		repo:  "flutter-template",
	},
	{
		name:  "dart-serverpod",
		owner: "zeabur",
		repo:  "serverpod-template",
	},

	/* deno */
	{
		name:  "deno-typescript",
		owner: "zeabur",
		repo:  "deno-typescript-template",
	},
	{
		name:  "deno-fresh",
		owner: "zeabur",
		repo:  "deno-fresh-template",
	},

	/* dotnet */
	{
		name:  "dotnet-aspnet-web-app",
		owner: "zeabur",
		repo:  "asp-dotnet-web-app-template",
	},
	{
		name:  "dotnet-aspnet-mvc",
		owner: "zeabur",
		repo:  "dotnet-mvc-template",
	},
	{
		name:  "dotnet-aspnet-web-api",
		owner: "zeabur",
		repo:  "asp-dotnet-web-api-template",
	},
	{
		name:  "dotnet-cli",
		owner: "zeabur",
		repo:  "dotnet-cli-template",
	},
	{
		name:  "dotnet-blazorwasm",
		owner: "zeabur",
		repo:  "blazorwasm-template",
	},

	/* elixir */
	{
		name:  "elixir-phoenix",
		owner: "zeabur",
		repo:  "elixir-phoenix-template",
	},
	{
		name:  "elixir-ecto",
		owner: "zeabur",
		repo:  "elixir-ecto-template",
	},

	/* gleam */
	{
		name:  "gleam",
		owner: "zeabur",
		repo:  "gleam-template",
	},

	/* go */
	{
		name:  "go-gin",
		owner: "zeabur",
		repo:  "gin-template",
	},

	/* java */
	{
		name:  "java-springboot-maven",
		owner: "zeabur",
		repo:  "spring-boot-maven-template",
	},
	{
		name:  "java-springboot-gradle",
		owner: "zeabur",
		repo:  "spring-boot-gradle-template",
	},

	/* node.js */
	{
		name:  "nodejs-payload",
		owner: "zeabur",
		repo:  "payload-template",
	},
	{
		name:  "nodejs-nextjs",
		owner: "zeabur",
		repo:  "nextjs-template",
	},
	{
		name:  "nodejs-nuxtjs",
		owner: "zeabur",
		repo:  "nuxtjs-template",
	},
	{
		name:  "node-sveltekit",
		owner: "zeabur",
		repo:  "svelte-kit-template",
	},
	{
		name:  "node-astro",
		owner: "zeabur",
		repo:  "astro-template",
	},
	{
		name:  "nodejs-starlight",
		owner: "zeabur",
		repo:  "starlight-template",
	},
	{
		name:  "nodejs-rspress",
		owner: "zeabur",
		repo:  "rspress-template",
	},
	{
		name:  "nodejs-vocs",
		owner: "zeabur",
		repo:  "vocs-template",
	},
	{
		name:  "nodejs-umi",
		owner: "zeabur",
		repo:  "umi-template",
	},
	{
		name:  "nodejs-remix",
		owner: "zeabur",
		repo:  "remix-template",
	},
	{
		name:  "nodejs-angular",
		owner: "zeabur",
		repo:  "angular-template",
	},
	{
		name:  "nodejs-zola",
		owner: "zeabur",
		repo:  "zola-template",
	},
	{
		name:  "nodejs-waku",
		owner: "zeabur",
		repo:  "waku-template",
	},
	{
		name:  "nodejs-vitepress",
		owner: "zeabur",
		repo:  "vitepress-template",
	},
	{
		name:  "nodejs-expressjs",
		owner: "zeabur",
		repo:  "expressjs-template",
	},
	{
		name:  "nodejs-nestjs",
		owner: "zeabur",
		repo:  "nestjs-template",
	},
	{
		name:  "nodejs-nuejs",
		owner: "zeabur",
		repo:  "nuejs-template",
	},
	{
		name:  "nodejs-foal",
		owner: "zeabur",
		repo:  "template-foal",
	},
	{
		name:  "nodejs-docusaurus",
		owner: "zeabur",
		repo:  "docusaurus-template",
	},
	{
		name:  "nodejs-slidev",
		owner: "zeabur",
		repo:  "slidev-template",
	},
	{
		name:  "nodejs-express-minio",
		owner: "zeabur",
		repo:  "express-minio-template",
	},
	{
		name:  "nodejs-qwik-city",
		owner: "zeabur",
		repo:  "qwik-city-template",
	},
	{
		name:  "nodejs-vite-vanilla",
		owner: "zeabur",
		repo:  "vite-vanilla-template",
	},
	{
		name:  "nodejs-sveltekit-v2",
		owner: "zeabur",
		repo:  "sveltekit-v2-template",
	},
	{
		name: "nodejs-a-lot-of-dependencies",
		dir:  "nodejs-a-lot-of-dependencies",
	},

	/* php */
	{
		name:  "php-thinkphp",
		owner: "zeabur",
		repo:  "thinkphp-template",
	},
	{
		name:  "php-codeigniter",
		owner: "zeabur",
		repo:  "codeigniter-template",
	},
	{
		name:  "php-laravel",
		owner: "zeabur",
		repo:  "laravel-template",
	},
	{
		name:  "php-symfony",
		owner: "zeabur",
		repo:  "symfony-template",
	},

	/* python */
	{
		name:  "python-django",
		owner: "zeabur",
		repo:  "django-template",
	},
	{
		name:  "python-django-static",
		owner: "zeabur",
		repo:  "django-static-template",
	},
	{
		name:  "python-django-static-whitenoise",
		owner: "testdrivenio",
		repo:  "django-static-media-files",
	},
	{
		name:  "python-flask",
		owner: "zeabur",
		repo:  "flask-template",
	},
	{
		name:  "python-flask-static",
		owner: "rtdtwo",
		repo:  "flask-static-tutorial",
	},
	{
		name:  "python-flask-mysql",
		owner: "zeabur",
		repo:  "flask-mysql-template",
	},
	{
		name:  "python-streamlit",
		owner: "zeabur",
		repo:  "streamlit-template",
	},
	{
		name:  "python-fastapi",
		owner: "zeabur",
		repo:  "fastapi-template",
	},
	{
		name:  "python-sanic",
		owner: "aragentum",
		repo:  "sanic-template",
	},
	{
		name: "python-hnswlib",
		dir:  "python-hnswlib",
	},
	{
		name: "python-zba651",
		dir:  "python-zba651",
	},

	/* ruby */
	{
		name:  "ruby-rails",
		owner: "zeabur",
		repo:  "rails-template",
	},

	/* rust */
	{
		name:  "rust-cli",
		owner: "zeabur",
		repo:  "rust-template",
	},
	{
		name:  "rust-axum",
		owner: "zeabur",
		repo:  "axum-template",
	},

	/* static */
	{
		name:  "static-html",
		owner: "schoolofdevops",
		repo:  "html-sample-app",
	},
	{
		name:  "static-mkdocs",
		owner: "zeabur",
		repo:  "mkdocs-template",
	},
	{
		name:  "static-hugo",
		owner: "zeabur",
		repo:  "hugo-template",
	},
	{
		name:  "static-hexo",
		owner: "zeabur",
		repo:  "hexo-template",
	},

	/* swift */
	{
		name:  "swift-vapor",
		owner: "zeabur",
		repo:  "vapor-template",
	},
}

func TestRealProjects(t *testing.T) {
	t.Parallel()

	pat := os.Getenv("GITHUB_PAT")

	for _, p := range projects {
		t.Run(p.name, func(t *testing.T) {
			var path string

			if p.dir != "" {
				path = p.dir
			} else {
				if pat == "" {
					t.Skip("GITHUB_PAT is not set")
				}

				path = fmt.Sprintf("https://github.com/%s/%s", p.owner, p.repo)
			}

			pt, pm := zeaburpack.Plan(zeaburpack.PlanOptions{
				SubmoduleName: &p.name,
				Path:          &path,
				AccessToken:   &pat,
			})
			Snapshot(t, p.name, pt, pm)
		})
	}
}

func Snapshot(t *testing.T, name string, pt types.PlanType, pm types.PlanMeta) {
	t.Helper()

	_ = os.Mkdir("snapshots", 0o755)

	snapshotContent := strings.Builder{}
	snapshotContent.WriteString("PlanType: " + string(pt) + "\n\n")
	snapshotContent.WriteString("Meta:\n")

	// sort meta first
	keys := maps.Keys(pm)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		snapshotContent.WriteString("  " + k + ": " + strconv.Quote(pm[k]) + "\n")
	}

	err := os.WriteFile(filepath.Join("snapshots", name+".txt"), []byte(snapshotContent.String()), 0o644)
	if err != nil {
		t.Error(err)
	}
}
