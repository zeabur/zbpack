package zeaburpack

import (
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/deno"
	"github.com/zeabur/zbpack/internal/golang"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/rust"
	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

type generateDockerfileOptions struct {
	src       afero.Fs
	HandleLog func(log string)
	planType  types.PlanType
	planMeta  types.PlanMeta
}

func generateDockerfile(opt *generateDockerfileOptions) (string, error) {
	dockerfile := ""
	planType := opt.planType
	planMeta := opt.planMeta
	src := opt.src

	switch planType {
	case types.PlanTypeDocker:

		dockerfileName := ""
		for _, filename := range []string{"dockerfile", "Dockerfile"} {
			if utils.HasFile(src, filename) {
				dockerfileName = filename
				break
			}
		}

		fileContent, err := afero.ReadFile(src, dockerfileName)
		if err != nil {
			return "", err
		}
		dockerfile = string(fileContent)
		return dockerfile, nil
	case types.PlanTypeNodejs:
		df, err := nodejs.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypeGo:
		df, err := golang.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypePython:
		df, err := python.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypeRuby:
		df, err := ruby.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypePHP:
		df, err := php.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypeJava:
		df, err := java.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypeDeno:
		df, err := deno.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case types.PlanTypeRust:
		df, err := rust.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	default:
		df, err := static.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	}

	return dockerfile, nil
}
