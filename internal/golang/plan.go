package golang

import (
	"bufio"
	"fmt"
	"path"

	"github.com/moznion/go-optional"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

type goPlanContext struct {
	plan.ProjectContext

	GoVersion optional.Option[string]
	Entry     optional.Option[string]

	Serverless optional.Option[bool]
}

func getGoVersion(ctx *goPlanContext) string {
	ver := &ctx.GoVersion
	if goVer, err := ver.Take(); err == nil {
		return goVer
	}

	fs := ctx.Source

	file, err := fs.Open("go.mod")
	if err != nil {
		return ""
	}
	defer func(file afero.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 3 && line[:3] == "go " {
			v := line[3:]
			*ver = optional.Some(v)
			return ver.Unwrap()
		}
	}

	*ver = optional.Some("1.18")
	return ver.Unwrap()
}

func getEntry(ctx *goPlanContext) string {
	ent := &ctx.Entry
	if entry, err := ent.Take(); err == nil {
		return entry
	}

	// in a basic go project, we assume the entrypoint is main.go in root directory
	if utils.HasFile(ctx.Source, "main.go") {
		*ent = optional.Some("")
		return ent.Unwrap()
	}

	// if there is no main.go in root directory, we assume it's a monorepo project.
	// in a general monorepo Go repo of service "user-service", the entry point might be `./cmd/user-service/main.go`
	entry := path.Join("cmd", ctx.SubmoduleName, "main.go")
	if utils.HasFile(ctx.Source, entry) {
		*ent = optional.Some(entry)
		return ent.Unwrap()
	}

	// We know it's a Go project, but we don't know how to build it.
	// We'll just return a generic Go plan type.
	*ent = optional.Some("")
	return ""
}

// GetMetaOptions is the options for GetMeta.
type GetMetaOptions struct {
	Src           afero.Fs
	Config        plan.ImmutableProjectConfiguration
	SubmoduleName string
}

func getServerless(ctx *goPlanContext) bool {
	return utils.GetExplicitServerlessConfig(ctx.Config).TakeOr(false)
}

// GetMeta gets the metadata of the Go project.
func GetMeta(opt GetMetaOptions) types.PlanMeta {
	ctx := &goPlanContext{
		ProjectContext: plan.ProjectContext{
			Source:        opt.Src,
			Config:        opt.Config,
			SubmoduleName: opt.SubmoduleName,
		},
	}
	meta := types.PlanMeta{}

	goVersion := getGoVersion(ctx)
	meta["goVersion"] = goVersion

	entry := getEntry(ctx)
	meta["entry"] = entry

	serverless := getServerless(ctx)
	if serverless {
		meta["serverless"] = "true"
	}

	return meta
}

func (i *identify) Planable(ctx plan.ProjectContext) bool {
	return utils.HasFile(ctx.Source, "go.mod")
}

func (i *identify) PlanAction(ctx plan.ProjectContext) (zbaction.Action, error) {
	pc := &goPlanContext{
		ProjectContext: ctx,
	}

	goVersion := getGoVersion(pc)
	entry := getEntry(pc)

	metadata := map[string]string{
		"goVersion": goVersion,
		"entry":     entry,
	}

	req := []zbaction.Requirement{
		{
			Expr:        fmt.Sprintf("matchVersion('go', '>= %s')", goVersion),
			Description: lo.ToPtr(fmt.Sprintf("Golang version must be greater than or equal to %s", goVersion)),
		},
	}

	job := []zbaction.Step{
		{
			Name: "Checkout sources",
			RunnableStep: zbaction.ProcStep{
				Uses: "zbpack/checkout",
			},
		},
		{
			Name: "Download dependencies",
			RunnableStep: zbaction.ProcStep{
				Uses: "zbpack/golang/mod-download",
				With: zbaction.ProcStepArgs{
					"optional": "true",
				},
			},
		},
		{
			ID:   "go-binary-step",
			Name: "Build the binary",
			RunnableStep: zbaction.ProcStep{
				Uses: "zbpack/golang/build",
				With: zbaction.ProcStepArgs{
					"entry": entry,
				},
			},
		},
		{
			ID:   "docker-image-step",
			Name: "Deploy the docker image",
			RunnableStep: zbaction.ProcStep{
				Uses: "zbpack/containerized",
				With: zbaction.ProcStepArgs{
					"context": "${out.go-binary-step.outDir}",
					"dockerfile": `
									FROM alpine
									COPY ./server /server
									CMD ["/server"]`,
				},
			},
		},
	}

	action := zbaction.Action{
		ID: "golang",
		Jobs: []zbaction.Job{
			{
				ID:    "build",
				Steps: job,
			},
		},
		Variables:    nil,
		Requirements: req,
		Metadata:     metadata,
	}

	return action, nil
}
