package action

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/pkg/action"
)

func init() {
	zbaction.RegisterProcedure("zbpack/containerized", func(args zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		ctx, ok := args["context"]
		if !ok {
			return nil, zbaction.NewErrRequiredArgument("context")
		}

		dockerfile, ok := args["dockerfile"]
		if !ok {
			return nil, zbaction.NewErrRequiredArgument("dockerfile")
		}

		cache, ok := args["cache"]
		if !ok {
			cache = "true"
		}

		return &ContainerizedProcedure{
			Context:    zbaction.NewArgumentStr(ctx),
			Dockerfile: zbaction.NewArgumentStr(dockerfile),
			Cache:      zbaction.NewArgumentBool(cache),
		}, nil
	})
}

// ContainerizedProcedure is a procedure that builds a Docker image based on the environment.
type ContainerizedProcedure struct {
	// Context is the directory to run the build in.
	Context zbaction.Argument[string]
	// Dockerfile is the content of the Dockerfile for runtime.
	Dockerfile zbaction.Argument[string]
	// Cache indicates whether to use cache when building the image.
	// By default, it is true.
	Cache zbaction.Argument[bool]
}

// Run builds a Docker image that determines if we should push and the image name based on environment.
func (c *ContainerizedProcedure) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	vc := sc.VariableContainer()
	cleanupStack := zbaction.CleanupStack{}
	cleanupFn := cleanupStack.WrapRun()

	args := zbaction.ProcStepArgs{
		"context":    c.Context.Value(sc.ExpandString),
		"dockerfile": c.Dockerfile.Value(sc.ExpandString),
		"cache":      strconv.FormatBool(c.Cache.Value(sc.ExpandString)),
	}

	push := checkShouldPush(vc)
	args["push"] = strconv.FormatBool(push)

	tag := findCustomTag(vc)
	if tag != "" {
		args["tag"] = tag
	} else if push {
		return cleanupFn, errors.New("cannot push without a custom tag")
	}

	// resolve a DockerArtifactAction
	procedure, err := zbaction.ResolveProcedure("action/artifact/docker", args)
	if err != nil {
		return cleanupFn, fmt.Errorf("resolve action/artifact/docker: %w", err)
	}

	// run the DockerArtifactAction
	dockerCleanupFn, err := procedure.Run(ctx, sc)
	cleanupStack.Push(dockerCleanupFn)
	if err != nil {
		return cleanupFn, fmt.Errorf("run action/artifact/docker: %w", err)
	}

	// If pushed, we are done.
	if push {
		return cleanupFn, nil
	}

	// If not pushed (local), we get the artifact TAR of this image and docker load it.
	artifact, ok := sc.GetThisOutput("artifact") /* see action/artifact/docker */
	if !ok {
		return cleanupFn, errors.New("no artifact found")
	}
	if _, ok := artifact.(string); !ok {
		return cleanupFn, errors.New("artifact is not a string")
	}

	artifactFile, err := os.Open(artifact.(string))
	if err != nil {
		return cleanupFn, fmt.Errorf("open artifact: %w", err)
	}
	defer func(artifactFile *os.File) {
		err := artifactFile.Close()
		if err != nil {
			fmt.Printf("close artifact file: %v\n", err)
		}
	}(artifactFile)

	cmd := exec.Command("docker", "load")
	cmd.Stdin = artifactFile
	cmd.Stdout = sc.Stdout()
	cmd.Stderr = sc.Stderr()
	cmd.Dir = sc.Root()

	err = cmd.Run()
	if err != nil {
		return cleanupFn, fmt.Errorf("load artifact: %w", err)
	}

	println("Docker image has been built & loaded locally.")

	return cleanupFn, nil
}

func checkShouldPush(vc zbaction.VariableContainer) bool {
	if v, ok := vc.GetVariable(action.ArgContainerPush); ok && v == "true" {
		return true
	}

	return false
}

func findCustomTag(vc zbaction.VariableContainer) string {
	if v, ok := vc.GetVariable(action.ArgContainerTag); ok {
		return v
	}

	return ""
}
