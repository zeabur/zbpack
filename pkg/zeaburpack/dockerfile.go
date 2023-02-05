package zeaburpack

import (
	_go "github.com/zeabur/zbpack/internal/go"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/ruby"
	"github.com/zeabur/zbpack/internal/plan"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/static"
	. "github.com/zeabur/zbpack/pkg/types"
	"os"
	"path"
)

type generateDockerfileOptions struct {
	SubmoduleName        string
	AbsPath              string
	HandleLog            func(log string)
	HandlePlanDetermined func(planType PlanType, planMeta PlanMeta)
}

func generateDockerfile(opt *generateDockerfileOptions) (string, error) {

	dockerfile := ""

	planner := plan.NewPlanner(opt.AbsPath, opt.SubmoduleName)
	planType, planMeta := planner.Plan()

	opt.HandlePlanDetermined(planType, planMeta)

	switch planType {
	case PlanTypeDocker:

		dockerfileName := ""
		if _, err := os.Stat(path.Join(opt.AbsPath, "dockerfile")); err == nil {
			dockerfileName = path.Join(opt.AbsPath, "dockerfile")
		} else {
			dockerfileName = path.Join(opt.AbsPath, "Dockerfile")
		}

		fileContent, err := os.ReadFile(dockerfileName)
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
	default:
		df, err := static.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	}

	return dockerfile, nil
}
