package nodejs

import (
	"strconv"
	"testing"

	"github.com/moznion/go-optional"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestGetNodeVersion_Empty(t *testing.T) {
	v := getNodeVersion("")
	assert.Equal(t, defaultNodeVersion, v)
}

func TestGetNodeVersion_Fixed(t *testing.T) {
	v := getNodeVersion("10")
	assert.Equal(t, "10", v)
}

func TestGetNodeVersion_Or(t *testing.T) {
	v := getNodeVersion("^10 || ^12 || ^14")
	assert.Equal(t, "14", v)
}

func TestGetNodeVersion_GreaterThanWithLessThan(t *testing.T) {
	v := getNodeVersion(">=16 <=20")
	assert.Equal(t, "20", v)
}

func TestGetNodeVersion_GreaterThan(t *testing.T) {
	v := getNodeVersion(">=4")
	assert.Equal(t, "4", v) // FIXME: should be the latest?
}

func TestGetNodeVersion_LessThan(t *testing.T) {
	v := getNodeVersion("<18")
	assert.Equal(t, "17", v)
}

func TestGetNodeVersion_Exact(t *testing.T) {
	v := getNodeVersion("16.0.0")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_Exact_WithEqualOp(t *testing.T) {
	v := getNodeVersion("=16.0.0")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_CaretMinor(t *testing.T) {
	v := getNodeVersion("^16.1.0")
	assert.Equal(t, "16", v)
}

func TestGetNodeVersion_TildeMinor(t *testing.T) {
	v := getNodeVersion("~16.0.1")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_ExactWithWildcard(t *testing.T) {
	v := getNodeVersion("16.0.*")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_TildeWithWildcard(t *testing.T) {
	v := getNodeVersion("~16.*")
	assert.Equal(t, "16", v)
}

func TestGetNodeVersion_NvmRcLts(t *testing.T) {
	v := getNodeVersion("lts/*")
	assert.Equal(t, strconv.FormatUint(maxLtsNodeVersion, 10), v)
}

func TestGetNodeVersion_NvmRcLatest(t *testing.T) {
	v := getNodeVersion("node")
	assert.Equal(t, strconv.FormatUint(maxNodeVersion, 10), v)
}

func TestGetNodeVersion_VPrefixedVersion(t *testing.T) {
	v := getNodeVersion("v20.11.0")
	assert.Equal(t, "20.11", v)
}

func TestGetInstallCmd_CustomizeInstallCmd(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")
	config.Set(plan.ConfigInstallCommand, "echo 'installed'")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}
	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// for customized installation command, no cache are allowed.
	assert.Contains(t, installlCmd, "COPY . .")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "echo 'installed'")
}

func TestGetInstallCmd_DefaultInstallCmd(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{}`), 0o644)
	_ = afero.WriteFile(src, "yarn.lock", []byte(``), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}

	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// for default installation command, cache is disabled.
	assert.NotContains(t, installlCmd, "COPY yarn.lock* .")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "yarn install")
}

func TestGetInstallCmd_CustomizeInstallCmdDeps(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{
	"dependencies": {
		"playwright-chromium": "*"
	}
}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")
	config.Set(plan.ConfigInstallCommand, "echo 'installed'")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}
	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// the playwright dependencies should be installed
	assert.Contains(t, installlCmd, "libnss3 libatk1.0-0 libatk-bridge2.0-0")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "echo 'installed'")
}

func TestGetMonorepoServiceRoot(t *testing.T) {
	t.Parallel()

	t.Run("pnpm-workspace", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*]`), 0o644)
		_ = afero.WriteFile(fs, "packages/service1/package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		serviceRoot := GetMonorepoAppRoot(ctx)
		assert.Equal(t, "packages/service1", serviceRoot)
	})

	t.Run("pnpm-workspace-two-glob", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*, apps/*]`), 0o644)
		_ = afero.WriteFile(fs, "apps/service1/package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		serviceRoot := GetMonorepoAppRoot(ctx)
		assert.Equal(t, "apps/service1", serviceRoot)
	})

	t.Run("yarn-workspace", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{"workspaces": ["packages/*"]}`), 0o644)
		_ = afero.WriteFile(fs, "packages/service1/package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		serviceRoot := GetMonorepoAppRoot(ctx)
		assert.Equal(t, "packages/service1", serviceRoot)
	})

	t.Run("config", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "services/service1/package.json", []byte(`{}`), 0o644)

		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigAppDir, "services/service1")

		ctx := &nodePlanContext{
			Src:                fs,
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
			Config:             config,
		}

		serviceRoot := GetMonorepoAppRoot(ctx)
		assert.Equal(t, "services/service1", serviceRoot)
	})
}

func TestNodePlanContext_GetServiceSource(t *testing.T) {
	t.Parallel()

	t.Run("generic", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{"main": "main.js"}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}
		fs, reldir := ctx.GetAppSource()

		assert.Equal(t, "", reldir)
		packageJSON, err := DeserializePackageJSON(fs)
		if assert.NoError(t, err) {
			assert.Equal(t, "main.js", packageJSON.Main)
		}
	})

	t.Run("monorepo", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*]`), 0o644)
		_ = afero.WriteFile(fs, "packages/service1/package.json", []byte(`{"main": "service1.js"}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}
		fs, reldir := ctx.GetAppSource()

		assert.Equal(t, "packages/service1", reldir)
		packageJSON, err := DeserializePackageJSON(fs)
		if assert.NoError(t, err) {
			assert.Equal(t, "service1.js", packageJSON.Main)
		}
	})
}

func TestNodePlanContext_GetServicePackageJSON(t *testing.T) {
	t.Parallel()

	t.Run("generic", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{"main": "main.js"}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}
		packageJSON := ctx.GetAppPackageJSON()
		assert.Equal(t, "main.js", packageJSON.Main)
	})

	t.Run("monorepo", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*]`), 0o644)
		_ = afero.WriteFile(fs, "packages/service1/package.json", []byte(`{"main": "service1.js"}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}
		packageJSON := ctx.GetAppPackageJSON()
		assert.Equal(t, "service1.js", packageJSON.Main)
	})
}

func TestInstallCommand(t *testing.T) {
	t.Parallel()

	t.Run("monorepo", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*]`), 0o644)
		_ = afero.WriteFile(fs, "packages/service1/package.json", []byte(`{"main": "service1.js"}`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/README", []byte("Hello, world!"), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		installCmd := GetInstallCmd(ctx)
		assert.Contains(t, installCmd, "COPY . .")
		assert.Contains(t, installCmd, "WORKDIR /src/packages/service1")
	})

	t.Run("normal (cache is disabled)", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		installCmd := GetInstallCmd(ctx)
		assert.Contains(t, installCmd, "COPY . .")
		assert.NotContains(t, installCmd, "WORKDIR")
	})

	t.Run("shouldCacheDependencies is false", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		config := plan.NewProjectConfigurationFromFs(fs, "")
		config.Set(ConfigCacheDependencies, false)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             config,
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		installCmd := GetInstallCmd(ctx)
		assert.Contains(t, installCmd, "COPY . .")
		assert.NotContains(t, installCmd, "WORKDIR")
	})
}

func TestGetStaticOutputDir(t *testing.T) {
	t.Run("vitepress, not specified docs directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"scripts": {
				"build": "vitepress build"
			},
			"devDependencies": {
				"vitepress": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		assert.Equal(t, ".vitepress/dist", GetStaticOutputDir(ctx))
	})

	t.Run("vitepress, specified docs directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"scripts": {
				"build": "vitepress build docs"
			},
			"devDependencies": {
				"vitepress": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		assert.Equal(t, "docs/.vitepress/dist", GetStaticOutputDir(ctx))
	})

	t.Run("vitepress, monorepo", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)
		_ = afero.WriteFile(fs, "pnpm-workspace.yaml", []byte(`packages: [packages/*]`), 0o644)
		_ = afero.WriteFile(fs, "packages/docs/package.json", []byte(`{
			"scripts": {
				"build": "pnpm -C ../ run build && vitepress build"
			},
			"devDependencies": {
				"vitepress": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		assert.Equal(t, ".vitepress/dist", GetStaticOutputDir(ctx))
	})
}

func TestGetStartCommand_Entry(t *testing.T) {
	t.Parallel()

	t.Run("node.js main", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"main": "hello.js"
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "node hello.js", startCmd)
	})

	t.Run("node.js fallback", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "node index.js", startCmd)
	})

	t.Run("bun main", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"main": "hello.js"
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
			Bun:                true,
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "bun hello.js", startCmd)
	})

	t.Run("bun fallback", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
			Bun:                true,
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "bun index.js", startCmd)
	})

	t.Run("serverless nitro", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"devDependencies": {
				"nitropack": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
			Framework:          optional.Some(types.NodeProjectFrameworkNitropack),
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "", startCmd)

		// bun
		ctx.StartCmd = optional.None[string]()
		ctx.Bun = true
		startCmd = GetStartCmd(ctx)
		assert.Equal(t, "", startCmd)
	})

	t.Run("containerized svelte", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"devDependencies": {
				"svelte": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:        fs,
			Config:     plan.NewProjectConfigurationFromFs(fs, ""),
			Framework:  optional.Some(types.NodeProjectFrameworkSvelte),
			Serverless: optional.Some(false),
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "node build/index.js", startCmd)
	})

	t.Run("containerized svelte with bun", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{
			"devDependencies": {
				"svelte": "*"
			}
		}`), 0o644)

		ctx := &nodePlanContext{
			Src:        fs,
			Config:     plan.NewProjectConfigurationFromFs(fs, ""),
			Framework:  optional.Some(types.NodeProjectFrameworkSvelte),
			Serverless: optional.Some(false),
			Bun:        true,
		}

		startCmd := GetStartCmd(ctx)
		assert.Equal(t, "bun build/index.js", startCmd)
	})

	for _, framework := range types.NitroBasedFrameworks {
		t.Run("nitro-"+string(framework), func(t *testing.T) {
			t.Run("nodejs", func(t *testing.T) {
				fs := afero.NewMemMapFs()
				_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

				ctx := &nodePlanContext{
					Src:                fs,
					Config:             plan.NewProjectConfigurationFromFs(fs, ""),
					ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
					Framework:          optional.Some(framework),
					Serverless:         optional.Some(false),
				}

				startCmd := GetStartCmd(ctx)
				assert.Equal(t, "HOST=0.0.0.0 node .output/server/index.mjs", startCmd)
			})

			t.Run("bun", func(t *testing.T) {
				fs := afero.NewMemMapFs()
				_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

				ctx := &nodePlanContext{
					Src:                fs,
					Config:             plan.NewProjectConfigurationFromFs(fs, ""),
					ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
					Bun:                true,
					Framework:          optional.Some(framework),
					Serverless:         optional.Some(false),
				}

				startCmd := GetStartCmd(ctx)
				assert.Equal(t, "HOST=0.0.0.0 bun .output/server/index.mjs", startCmd)
			})
		})
	}
}

func TestGetStartCommand_Config(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{
			"main": "hello.js"
		}`), 0o644)
	config := plan.NewProjectConfigurationFromFs(fs, "")

	config.Set(plan.ConfigStartCommand, "echo 'hello'")

	ctx := &nodePlanContext{
		Src:                fs,
		Config:             config,
		ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
	}

	startCmd := GetStartCmd(ctx)
	assert.Equal(t, "echo 'hello'", startCmd)
}

func TestGetServerless(t *testing.T) {
	t.Parallel()

	for _, nitroFramework := range types.NitroBasedFrameworks {
		t.Run("nitro-"+string(nitroFramework), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

			ctx := &nodePlanContext{
				Src:                fs,
				Config:             plan.NewProjectConfigurationFromFs(fs, ""),
				ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
				Framework:          optional.Some(nitroFramework),
			}

			assert.True(t, getServerless(ctx))
		})
	}

	t.Run("unknown", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "package.json", []byte(`{}`), 0o644)

		ctx := &nodePlanContext{
			Src:                fs,
			Config:             plan.NewProjectConfigurationFromFs(fs, ""),
			ProjectPackageJSON: lo.Must(DeserializePackageJSON(fs)),
			Framework:          optional.Some[types.NodeProjectFramework](""),
		}

		assert.False(t, getServerless(ctx))
	})
}
