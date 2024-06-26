package rust

import (
	"log"
	"os"
	"strings"

	"github.com/moznion/go-optional"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type rustPlanContext struct {
	Src    afero.Fs
	Config plan.ImmutableProjectConfiguration

	SubmoduleName string
	Serverless    optional.Option[bool]
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src afero.Fs

	Config plan.ImmutableProjectConfiguration
	// In Rust, the submodule name is the binary name.
	SubmoduleName string
}

// getServerless gets the serverless flag from the configuration.
func getServerless(ctx *rustPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
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

// GetMeta gets the metadata of the Rust project.
func GetMeta(options GetMetaOptions) types.PlanMeta {
	ctx := &rustPlanContext{
		Src:           options.Src,
		SubmoduleName: options.SubmoduleName,
		Config:        options.Config,
	}

	var opensslFlag string
	if needOpenssl(ctx.Src) {
		opensslFlag = "yes"
	} else {
		opensslFlag = "no"
	}

	meta := types.PlanMeta{}

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = "true"
	}

	meta["BinName"] = ctx.SubmoduleName
	meta["NeedOpenssl"] = opensslFlag

	return meta
}
