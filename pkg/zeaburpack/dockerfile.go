package zeaburpack

import (
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/deno"
	_go "github.com/zeabur/zbpack/internal/go"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/rust"
	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/internal/utils"
	. "github.com/zeabur/zbpack/pkg/types"
)

type generateDockerfileOptions struct {
	src       afero.Fs
	HandleLog func(log string)
	planType  PlanType
	planMeta  PlanMeta
}

func generateDockerfile(opt *generateDockerfileOptions) (string, error) {
	dockerfile := ""
	planType := opt.planType
	planMeta := opt.planMeta
	src := opt.src

	switch planType {
	case PlanTypeDocker:

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
	case PlanTypeNodejs:
		df, err := nodejs.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeGo:
		df, err := _go.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypePython:
		df, err := python.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeRuby:
		df, err := ruby.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypePhp:
		df, err := php.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeJava:
		df, err := java.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeDeno:
		df, err := deno.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeRust:
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
