package zeaburpack

import (
	"fmt"
	_go "github.com/zeabur/zbpack/internal/go"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/php"
	"github.com/zeabur/zbpack/internal/plan"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/static"
	. "github.com/zeabur/zbpack/pkg/types"
	"os"
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
		fileContent, err := os.ReadFile("dockerfile")
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
		err := fmt.Errorf("ruby is not supported yet")
		return "", err
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
