package zeaburpack

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/zeabur/zbpack/internal/plan"
	"github.com/zeabur/zbpack/internal/source"

	. "github.com/zeabur/zbpack/pkg/types"
)

type BuildOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// HandlePlanDetermined is a callback function that will be called when
	// the build plan is determined.
	HandlePlanDetermined *func(PlanType, PlanMeta)

	// HandleBuildFailed is a callback function that will be called when
	// the build failed.
	HandleBuildFailed *func(error)

	// HandleLog is a function that will be called when a log is emitted.
	HandleLog *func(string)

	// Path is the path to the project directory.
	Path *string

	// ResultImage is the name of the image that will be built.
	ResultImage *string

	// UserVars is a map of user variables that will be used in the Dockerfile.
	UserVars *map[string]string

	// Interactive is a flag to indicate if the build should be interactive.
	Interactive *bool

	// CustomBuildCommand is a custom build command that will be used instead of the default one.
	CustomBuildCommand *string

	// CustomStartCommand is a custom start command that will be used instead of the default one.
	CustomStartCommand *string

	// OutputDir is the directory where the build artifacts will be stored.
	// Once provided, the service will deploy as static files with nginx.
	OutputDir *string

	CacheFrom *string
}

// Build will analyze the project, determine the plan and build the image.
func Build(opt *BuildOptions) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if opt.Path == nil || *opt.Path == "" {
		opt.Path = &wd
	} else if !strings.HasPrefix(*opt.Path, "/") {
		p := path.Join(wd, *opt.Path)
		opt.Path = &p
	}

	if opt.SubmoduleName == nil {
		emptySubmoduleName := ""
		opt.SubmoduleName = &emptySubmoduleName
	}

	if opt.ResultImage == nil || *opt.ResultImage == "" {
		img := path.Base(*opt.Path)
		opt.ResultImage = &img
	}

	*opt.ResultImage = strings.ToLower(*opt.ResultImage)
	*opt.ResultImage = strings.ReplaceAll(*opt.ResultImage, "_", "-")

	if opt.UserVars == nil {
		emptyUserVars := make(map[string]string)
		opt.UserVars = &emptyUserVars
	}

	var handleLog func(log string)
	if opt.HandleLog == nil {
		handleLog = func(log string) {
			println(log)
		}
	} else {
		handleLog = func(log string) {
			(*opt.HandleLog)(log)
			println(log)
		}
	}

	var handleBuildFailed func(error)
	if opt.HandleBuildFailed == nil {
		handleBuildFailed = func(err error) {
			println("Build failed: " + err.Error())
		}
	} else {
		handleBuildFailed = func(err error) {
			println("Build failed: " + err.Error())
			(*opt.HandleBuildFailed)(err)
		}
	}

	if strings.HasPrefix(*opt.Path, "https://") {
		println("Build from git repository is not supported yet")
		handleBuildFailed(nil)
		return fmt.Errorf("build from git repository is not supported yet")
	}

	src := source.NewLocalSource(*opt.Path)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             &src,
			SubmoduleName:      *opt.SubmoduleName,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
		},
	)

	t, m := planner.Plan()

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			src:       &src,
			HandleLog: handleLog,
			planType:  t,
			planMeta:  m,
		},
	)
	if err != nil {
		println("Failed to generate Dockerfile: " + err.Error())
		handleBuildFailed(err)
		return err
	}

	// If the build is interactive, we will print the Dockerfile to the console.
	buildImageHandleLog := &handleLog
	if opt.Interactive != nil && *opt.Interactive {
		buildImageHandleLog = nil
	}

	err = buildImage(
		&BuildImageOptions{
			Dockerfile:          dockerfile,
			AbsPath:             *opt.Path,
			UserVars:            *opt.UserVars,
			ResultImage:         *opt.ResultImage,
			HandleLog:           buildImageHandleLog,
			PlainDockerProgress: opt.Interactive == nil || !*opt.Interactive,
			CacheFrom:           opt.CacheFrom,
		},
	)
	if err != nil {
		println("Failed to build image: " + err.Error())
		handleBuildFailed(err)
		return err
	}

	// If the dockerfile is using Nginx as a runtime, we will copy the static files to the output directory.
	if strings.Contains(dockerfile, "nginx:alpine as runtime") {
		err = extractStaticOutput(*opt.ResultImage, opt)
		if err != nil {
			println("Failed to copy static files: " + err.Error())
		}
	}

	if opt.Interactive != nil && *opt.Interactive {
		handleLog("\n\033[32mBuild successful\033[0m\n")
		handleLog("\033[90m" + "To run the image, use the following command:" + "\033[0m")
		handleLog("docker run -p 8080:8080 -it " + *opt.ResultImage)
	}

	return nil
}

func extractStaticOutput(resultImage string, opt *BuildOptions) error {
	copyFiles := `FROM ` + resultImage + `
CMD ["cp", "-r", "/usr/share/nginx/html/static", "/out/"]`

	tempDir := os.TempDir()
	buildID := strconv.Itoa(rand.Int())

	err := os.MkdirAll(path.Join(tempDir, buildID), 0o755)
	if err != nil {
		return err
	}

	defer func() {
		err = os.RemoveAll(path.Join(tempDir, buildID))
		if err != nil {
			println("\033[31m" + "Failed to remove temp directory" + "\033[0m")
			println("\033[31m" + err.Error() + "\033[0m")
		}
	}()

	dfPath := path.Join(tempDir, buildID, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(copyFiles), 0o644); err != nil {
		return err
	}

	args := []string{"build", "-t", "copy", "-f", dfPath, *opt.Path}
	err = exec.Command("docker", args...).Run()
	if err != nil {
		return err
	}

	defer func() {
		cmd := exec.Command("docker", "rmi", "copy")
		err = cmd.Run()
		if err != nil {
			println("\033[31m" + "Failed to remove copy image" + "\033[0m")
			println("\033[31m" + err.Error() + "\033[0m")
		}
	}()

	hostPath := *opt.Path + "/.zeabur/output"
	containerPath := "/out"
	v := hostPath + ":" + containerPath

	println("docker", "run", "--rm", "-v", v, "copy")
	err = exec.Command("docker", "run", "--rm", "-v", v, "copy").Run()
	if err != nil {
		println("\033[31m" + "Failed to copy files to .zeabur/output/static" + "\033[0m")
		println("\033[31m" + err.Error() + "\033[0m")
	}

	return nil
}
