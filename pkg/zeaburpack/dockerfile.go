package zeaburpack

import (
	"fmt"
	_go "github.com/zeabur/zbpack/internal/go"
	"github.com/zeabur/zbpack/internal/java"
	"github.com/zeabur/zbpack/internal/nodejs"
	"github.com/zeabur/zbpack/internal/plan"
	"github.com/zeabur/zbpack/internal/python"
	"github.com/zeabur/zbpack/internal/static"
	. "github.com/zeabur/zbpack/pkg/types"
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
		opt.HandleLog("Using PlanTypeDocker to build image.")
		return "", nil
	case PlanTypeNodejs:
		opt.HandleLog("Using PlanTypeNodejs to build image.")
		df, err := nodejs.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeGo:
		opt.HandleLog("Using PlanTypeGo to build image.")
		df, err := _go.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypePython:
		opt.HandleLog("Using PlanTypePython to build image.")
		df, err := python.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	case PlanTypeRuby:
		opt.HandleLog("Using PlanTypeRuby to build image.")
		err := fmt.Errorf("ruby is not supported yet")
		return "", err
	case PlanTypePhp:
		opt.HandleLog("Using PlanTypePhp to build image.")
		err := fmt.Errorf("php is not supported yet")
		return "", err
	case PlanTypeJava:
		opt.HandleLog("Using PlanTypeJava to build image.")
		df, err := java.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	default:
		opt.HandleLog("Using PlanTypeStatic to build image.")
		df, err := static.GenerateDockerfile(planMeta)
		if err != nil {
			return "", err
		}
		dockerfile = df
	}

	return dockerfile, nil
}
