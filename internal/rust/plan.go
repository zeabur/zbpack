package rust

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// ConfigRustEntry is the key for the binary entry name of the application.
//
// If this key is not set, the default value is the binary with the submodule name.
// If there is no such submodule, it picks the first binary it found.
const ConfigRustEntry = "rust.entry"

// ConfigRustAppDir is the key for the directory of the application.
//
// If this key is not set, the default value is the current directory – "/".
const ConfigRustAppDir = "rust.app_dir"

// ConfigRustAssets is the key for the assets of the application.
// It is an array.
//
// The assets will be copied to the root of the application.
const ConfigRustAssets = "rust.assets"

// ConfigPreStartCommand is the key for the command before `CMD`.
// Useful for installing dependencies for runtime.
const ConfigPreStartCommand = "pre_start_command"

type rustPlanContext struct {
	Src           afero.Fs
	Config        plan.ImmutableProjectConfiguration
	SubmoduleName string
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration
	// In Rust, the submodule name is the binary name.
	SubmoduleName string
}

// getServerless gets the serverless flag from the configuration.
func getServerless(ctx *rustPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
}

// getEntry gets the entry name from the configuration.
func getEntry(ctx *rustPlanContext) string {
	if entry, err := plan.Cast(ctx.Config.Get(ConfigRustEntry), cast.ToStringE).Take(); err == nil {
		return entry
	}

	if ctx.SubmoduleName != "" {
		return ctx.SubmoduleName
	}

	// If there is no entry named 'main', we find
	// the first binary in the artifact directory.
	return "main"
}

// getAppDir gets the application directory from the configuration.
func getAppDir(ctx *rustPlanContext) string {
	if appDir, err := plan.Cast(ctx.Config.Get(ConfigRustAppDir), cast.ToStringE).Take(); err == nil {
		if appDir == "/" {
			return "."
		}

		return appDir
	}

	return "." // current directory relative to root.
}

// getAssets gets the assets list that needs to copy from project directory.
func getAssets(ctx *rustPlanContext) []string {
	assets := plan.Cast(ctx.Config.Get(ConfigRustAssets), cast.ToStringSliceE).TakeOr([]string{})
	if len(assets) != 0 {
		return assets
	}

	// Legacy configuration.
	zeaburPreserve, err := afero.ReadFile(ctx.Src, ".zeabur-preserve")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}

		return assets
	}

	return strings.FieldsFunc(string(zeaburPreserve), func(r rune) bool { return r == '\n' })
}

// needOpenssl checks if the project needs openssl.
func needOpenssl(source afero.Fs) bool {
	for _, file := range []string{"Cargo.toml", "Cargo.lock"} {
		file, err := utils.ReadFileToUTF8(source, file)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Println(err)
			}
			continue
		}

		if strings.Contains(string(file), "openssl") {
			return true
		}
	}
	return false
}

func getBuildCommand(ctx *rustPlanContext) string {
	return plan.Cast(ctx.Config.Get(plan.ConfigBuildCommand), cast.ToStringE).TakeOr("")
}

func getStartCommand(ctx *rustPlanContext) string {
	return plan.Cast(ctx.Config.Get(plan.ConfigStartCommand), cast.ToStringE).TakeOr("")
}

func getPreStartCommand(ctx *rustPlanContext) string {
	return plan.Cast(ctx.Config.Get(ConfigPreStartCommand), cast.ToStringE).TakeOr("")
}

// GetMeta gets the metadata of the Rust project.
func GetMeta(options GetMetaOptions) types.PlanMeta {
	ctx := &rustPlanContext{
		Src:           options.Src,
		SubmoduleName: options.SubmoduleName,
		Config:        options.Config,
	}

	meta := types.PlanMeta{
		"openssl":    strconv.FormatBool(needOpenssl(ctx.Src)),
		"serverless": strconv.FormatBool(getServerless(ctx)),
		"entry":      getEntry(ctx),
		"appDir":     getAppDir(ctx),
		// assets/1:assets/2:...
		"assets":          strings.Join(getAssets(ctx), ":"),
		"buildCommand":    getBuildCommand(ctx),
		"startCommand":    getStartCommand(ctx),
		"preStartCommand": getPreStartCommand(ctx),
	}

	return meta
}
