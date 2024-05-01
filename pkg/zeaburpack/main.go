package zeaburpack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/codeclysm/extract/v3"
	cp "github.com/otiai10/copy"
	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/nodejs/nextjs"
	"github.com/zeabur/zbpack/internal/nodejs/nuxtjs"
	"github.com/zeabur/zbpack/internal/nodejs/remix"
	"github.com/zeabur/zbpack/internal/nodejs/waku"
	"github.com/zeabur/zbpack/internal/static"
	"github.com/zeabur/zbpack/pkg/plan"
	"github.com/zeabur/zbpack/pkg/types"
)

// BuildOptions is the options for the Build function.
type BuildOptions struct {
	// SubmoduleName is the of the submodule to build.
	// For example, if directory is considered as a Go project,
	// submoduleName would be used to try file in `cmd` directory.
	// in Zeabur internal system, this is the name of the service.
	SubmoduleName *string

	// HandlePlanDetermined is a callback function that will be called when
	// the build plan is determined.
	HandlePlanDetermined *func(types.PlanType, types.PlanMeta)

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
	CacheTo   *string

	// ProxyRegistry is the registry to be used for the image.
	// See referenceConstructor for more details.
	ProxyRegistry *string

	// PushImage is a flag to indicate if the image should be pushed to the registry.
	PushImage bool
}

// Build will analyze the project, determine the plan and build the image.
func Build(opt *BuildOptions) error {

	// clean up the buildkit output directory after the build
	defer func() {
		_ = os.RemoveAll(path.Join(os.TempDir(), "zbpack/buildkit"))
	}()

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

	src := afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	UpdateOptionsOnConfig(opt, config)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:             src,
			Config:             config,
			SubmoduleName:      submoduleName,
			CustomBuildCommand: opt.CustomBuildCommand,
			CustomStartCommand: opt.CustomStartCommand,
			OutputDir:          opt.OutputDir,
		},
		SupportedIdentifiers(config)...,
	)

	t, m := planner.Plan()

	PrintPlanAndMeta(t, m, handleLog)

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
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

	// Remove .zeabur directory if exists
	_ = os.RemoveAll(path.Join(*opt.Path, ".zeabur"))

	err = buildImage(
		&buildImageOptions{
			PlanType: t,
			PlanMeta: m,

			Dockerfile:          dockerfile,
			AbsPath:             *opt.Path,
			UserVars:            *opt.UserVars,
			HandleLog:           buildImageHandleLog,
			PlainDockerProgress: opt.Interactive == nil || !*opt.Interactive,

			ResultImage: *opt.ResultImage,
			PushImage:   opt.PushImage,

			CacheFrom:     opt.CacheFrom,
			CacheTo:       opt.CacheTo,
			ProxyRegistry: opt.ProxyRegistry,
		},
	)
	if err != nil {
		println("Failed to build image: " + err.Error())
		handleBuildFailed(err)
		return err
	}

	dockerBuildOutput := path.Join(os.TempDir(), "zbpack/buildkit")
	// decompress TAR to the output directory
	func() {
		if err := os.MkdirAll(dockerBuildOutput, 0755); err != nil {
			println("Failed to create output directory: " + err.Error())
			return
		}

		// decompress the given TAR file to the output directory
		tarFile, err := os.Open(ServerlessTarPath)
		if err != nil {
			if m["serverless"] == "true" {
				println("Failed to open TAR: " + err.Error())
			}
			return
		}
		defer func(tarFile *os.File) {
			_ = tarFile.Close()

			// clean up TAR file
			_ = os.Remove(ServerlessTarPath)
		}(tarFile)

		err = extract.Tar(context.TODO(), tarFile, dockerBuildOutput, func(filename string) string {
			switch filename {
			case ".git", ".github", ".vscode", ".idea", ".gitignore",
				"Dockerfile", "LICENSE", "README.md", "Makefile",
				".pre-commit-config.yaml":
				return "" // skip these files
			default:
				return filename
			}
		})
		if err != nil {
			println("Failed to decompress TAR: " + err.Error())
			return
		}
	}()

	dotZeaburDirInOutput := path.Join(dockerBuildOutput, ".zeabur")

	stat, err := os.Stat(dotZeaburDirInOutput)
	if err == nil && stat.IsDir() {
		_ = os.MkdirAll(path.Join(*opt.Path, ".zeabur"), 0755)
		err = cp.Copy(dotZeaburDirInOutput, path.Join(*opt.Path, ".zeabur"))
		if err != nil {
			println("Failed to copy .zeabur directory from the output: " + err.Error())
		}
	}

	if t == types.PlanTypeGo && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), path.Join(*opt.Path, ".zeabur/output/functions/__go.func"))
		if err != nil {
			println("Failed to copy serverless function: " + err.Error())
		}

		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__go.func"))
		if err != nil {
			handleLog("Failed to write function config to \".zeabur/output/functions/__go.func\": " + err.Error())
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__go"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypeRust && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), path.Join(*opt.Path, ".zeabur/output/functions/__rs.func"))
		if err != nil {
			println("Failed to copy serverless function: " + err.Error())
		}

		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__rs.func"))
		if err != nil {
			handleLog("Failed to write function config to \".zeabur/output/functions/__rs.func\": " + err.Error())
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__rs"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypePython && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		funcPath := path.Join(*opt.Path, ".zeabur/output/functions/__py.func")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), funcPath)
		if err != nil {
			println("Failed to copy serverless function: " + err.Error())
		}

		// if there is "static" directory in the output, we will copy it to .zeabur/output/static
		statStatic, errStatic := os.Stat(path.Join(os.TempDir(), "/zbpack/buildkit/static"))
		if errStatic == nil && statStatic.IsDir() {
			err = cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit/static"), path.Join(*opt.Path, ".zeabur/output/static"))
			if err != nil {
				println("Failed to copy static directory: " + err.Error())
			}
		}

		var venvPath string
		dirs, err := os.ReadDir(funcPath)
		if err == nil {
			for _, dir := range dirs {
				if !dir.IsDir() {
					continue
				}
				readLib, err := os.Stat(path.Join(funcPath, dir.Name(), "lib", "python"+m["pythonVersion"], "site-packages"))
				if err != nil || !readLib.IsDir() {
					continue
				}
				venvPath = path.Join(funcPath, dir.Name())
			}
		}

		if venvPath != "" {
			oldSp := path.Join(*opt.Path, ".zeabur/output/functions/__py.func/.site-packages")
			newSp := path.Join(venvPath, "lib", "python"+m["pythonVersion"], "site-packages")
			_ = os.RemoveAll(oldSp)
			_ = cp.Copy(newSp, oldSp)
			_ = os.RemoveAll(venvPath)
		}

		pythonVersionWithoutDot := strings.ReplaceAll(m["pythonVersion"], ".", "")
		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "python" + pythonVersionWithoutDot}
		if m["entry"] != "" {
			funcConfig.Entry = m["entry"]
		}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__py.func"))
		if err != nil {
			handleLog("Failed to write function config to \".zeabur/output/functions/__py.func\": " + err.Error())
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__py"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkWaku) && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err = waku.TransformServerless(*opt.Path)
		if err != nil {
			log.Println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkNextJs) && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err = nextjs.TransformServerless(*opt.Path)
		if err != nil {
			log.Println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkRemix) && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err = remix.TransformServerless(*opt.Path)
		if err != nil {
			log.Println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkNuxtJs) && m["serverless"] == "true" {
		println("Transforming build output to serverless format ...")
		err = nuxtjs.TransformServerless(*opt.Path)
		if err != nil {
			log.Println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if m["outputDir"] != "" {
		println("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.Path, m)
		if err != nil {
			println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeDart && m["framework"] == string(types.DartFrameworkFlutter) {
		println("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.Path, m)
		if err != nil {
			println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if t == types.PlanTypeStatic {
		println("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.Path, m)
		if err != nil {
			println("Failed to transform serverless: " + err.Error())
			handleBuildFailed(err)
			return err
		}
	}

	if opt.Interactive != nil && *opt.Interactive {
		handleLog("\n\033[32mBuild successful\033[0m\n")
		if m["serverless"] == "true" {
			handleLog("\033[90m" + "The compiled serverless function has been saved in the .zeabur directory." + "\033[0m")
		} else {
			handleLog("\033[90m" + "To run the image, use the following command:" + "\033[0m")
			if (t == types.PlanTypeNodejs && m["outputDir"] != "") || t == types.PlanTypeStatic {
				handleLog("npx serve .zeabur/output/static")
			} else {
				handleLog("docker run -p 8080:8080 -it " + *opt.ResultImage)
			}
		}
	}

	return nil
}
