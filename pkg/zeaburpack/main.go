package zeaburpack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	"github.com/zeabur/zbpack/internal/utils"
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

	// LogWriter is a [io.Writer] that will be written when a log is emitted.
	// nil to use the default log writer.
	LogWriter io.Writer

	// Path is the path to the project directory.
	Path *string

	// ResultImage is the name of the image that will be built.
	ResultImage *string

	// UserVars is a map of user variables that will be used in the Dockerfile.
	UserVars *map[string]string

	// Interactive is a flag to indicate if the build should be interactive.
	Interactive *bool

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

	if opt.LogWriter == nil {
		opt.LogWriter = os.Stderr
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

	if strings.HasPrefix(*opt.Path, "https://") {
		opt.Log("Build from git repository is not supported yet\n")
		return fmt.Errorf("build from git repository is not supported yet")
	}

	src := afero.NewBasePathFs(afero.NewOsFs(), *opt.Path)
	submoduleName := lo.FromPtrOr(opt.SubmoduleName, "")
	config := plan.NewProjectConfigurationFromFs(src, submoduleName)

	planner := plan.NewPlanner(
		&plan.NewPlannerOptions{
			Source:        src,
			Config:        config,
			SubmoduleName: submoduleName,
		},
		SupportedIdentifiers(config)...,
	)

	t, m := planner.Plan()

	PrintPlanAndMeta(t, m, opt.LogWriter)

	if opt.HandlePlanDetermined != nil {
		(*opt.HandlePlanDetermined)(t, m)
	}

	dockerfile, err := generateDockerfile(
		&generateDockerfileOptions{
			planType: t,
			planMeta: m,
		},
	)
	if err != nil {
		opt.Log("Failed to generate Dockerfile: %s\n", err)
		return err
	}

	// Remove .zeabur directory if exists
	_ = os.RemoveAll(path.Join(*opt.Path, ".zeabur"))

	// Inject dockerfile to contain the variables, registry, etc.
	newDockerfile := InjectDockerfile(dockerfile, opt.ProxyRegistry, *opt.UserVars)

	err = buildImage(
		&buildImageOptions{
			PlanType: t,
			PlanMeta: m,

			Dockerfile:          newDockerfile,
			AbsPath:             *opt.Path,
			UserVars:            *opt.UserVars,
			PlainDockerProgress: opt.Interactive == nil || !*opt.Interactive,

			ResultImage: *opt.ResultImage,
			PushImage:   opt.PushImage,

			CacheFrom: opt.CacheFrom,
			CacheTo:   opt.CacheTo,

			LogWriter: opt.LogWriter,
		},
	)
	if err != nil {
		opt.Log("Failed to build image: %s\n", err)
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
				opt.Log("Failed to open TAR file: %s\n", err)
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
			opt.Log("Failed to decompress TAR: %s\n", err)
			return
		}
	}()

	dotZeaburDirInOutput := path.Join(dockerBuildOutput, ".zeabur")

	stat, err := os.Stat(dotZeaburDirInOutput)
	if err == nil && stat.IsDir() {
		_ = os.MkdirAll(path.Join(*opt.Path, ".zeabur"), 0o755)
		err = cp.Copy(dotZeaburDirInOutput, path.Join(*opt.Path, ".zeabur"))
		if err != nil {
			opt.Log("Failed to copy .zeabur directory from the output: %s\n", err)
		}
	}

	if t == types.PlanTypeNix {
		dockerTarName := filepath.Join(dockerBuildOutput, "result")

		if !opt.PushImage {
			// SAFE: zbpack are managed by ourselves. Besides,
			// macOS does not contain policy.json by default.
			skopeoCmd := exec.Command("skopeo", "copy", "--insecure-policy", "docker-archive:"+dockerTarName, "docker-daemon:"+*opt.ResultImage+":latest")
			skopeoCmd.Stdout = opt.LogWriter
			skopeoCmd.Stderr = opt.LogWriter
			if err := skopeoCmd.Run(); err != nil {
				return fmt.Errorf("run skopeo copy: %w", err)
			}
		} else {
			// SAFE: zbpack are managed by ourselves. Besides,
			// macOS does not contain policy.json by default.
			skopeoCmd := exec.Command("skopeo", "copy", "--insecure-policy", "docker-archive:"+dockerTarName, "docker://"+*opt.ResultImage)
			skopeoCmd.Stdout = opt.LogWriter
			skopeoCmd.Stderr = opt.LogWriter
			if err := skopeoCmd.Run(); err != nil {
				return fmt.Errorf("run skopeo copy: %w", err)
			}
		}

		// remove the TAR since we have imported it
		_ = os.Remove(dockerTarName)
	}

	if t == types.PlanTypeGo && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), path.Join(*opt.Path, ".zeabur/output/functions/__go.func"))
		if err != nil {
			opt.Log("Failed to copy serverless function: %s\n", err)
		}

		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__go.func"))
		if err != nil {
			opt.Log("Failed to write function config to \".zeabur/output/functions/__go.func\": %s\n", err)
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__go"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0o644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypeRust && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), path.Join(*opt.Path, ".zeabur/output/functions/__rs.func"))
		if err != nil {
			opt.Log("Failed to copy serverless function: %s\n", err)
		}

		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "binary", Entry: "./main"}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__rs.func"))
		if err != nil {
			opt.Log("Failed to write function config to \".zeabur/output/functions/__rs.func\": %s\n", err)
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__rs"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0o644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypePython && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		funcPath := path.Join(*opt.Path, ".zeabur/output/functions/__py.func")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), funcPath)
		if err != nil {
			opt.Log("Failed to copy serverless function: %s\n", err)
		}

		// if there is "static" directory in the output, we will copy it to .zeabur/output/static
		statStatic, errStatic := os.Stat(path.Join(os.TempDir(), "/zbpack/buildkit/static"))
		if errStatic == nil && statStatic.IsDir() {
			err = cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit/static"), path.Join(*opt.Path, ".zeabur/output/static"))
			if err != nil {
				opt.Log("Failed to copy static directory: %s\n", err)
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
			opt.Log("Failed to write function config to \".zeabur/output/functions/__py.func\": %s\n", err)
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__py"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0o644)
		if err != nil {
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkWaku) && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err = waku.TransformServerless(*opt.Path)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkNextJs) && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err = nextjs.TransformServerless(*opt.Path)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if t == types.PlanTypeNodejs && m["framework"] == string(types.NodeProjectFrameworkRemix) && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err = remix.TransformServerless(*opt.Path)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if (t == types.PlanTypeNodejs || t == types.PlanTypeBun) && types.IsNitroBasedFramework(m["framework"]) && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err = nuxtjs.TransformServerless(*opt.Path)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if t == types.PlanTypeGleam && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")

		funcPath := path.Join(*opt.Path, ".zeabur/output/functions/__erl.func")
		err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), funcPath)
		if err != nil {
			opt.Log("Failed to copy serverless function: %s\n", err)
		}

		content, err := os.ReadFile(path.Join(*opt.Path, ".zeabur/output/functions/__erl.func/entrypoint.sh"))
		if err != nil {
			opt.Log("Failed to read entrypoint.sh: %s\n", err)
		}

		entry := utils.ExtractErlangEntryFromGleamEntrypointShell(string(content))
		funcConfig := types.ZeaburOutputFunctionConfig{Runtime: "erlang27", Entry: entry}

		err = funcConfig.WriteTo(path.Join(*opt.Path, ".zeabur/output/functions/__erl.func"))
		if err != nil {
			opt.Log("Failed to write function config to \".zeabur/output/functions/__erl.func\": %s\n", err)
		}

		config := types.ZeaburOutputConfig{Routes: []types.ZeaburOutputConfigRoute{{Src: ".*", Dest: "/__erl"}}}

		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(*opt.Path, ".zeabur/output/config.json"), configBytes, 0o644)
		if err != nil {
			return err
		}
	}

	if m["outputDir"] != "" {
		opt.Log("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.Path, m)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if t == types.PlanTypeStatic && m["serverless"] == "true" {
		opt.Log("Transforming build output to serverless format ...")
		err = static.TransformServerless(*opt.Path, m)
		if err != nil {
			opt.Log("Failed to transform serverless: %s\n", err)
			return err
		}
	}

	if opt.Interactive != nil && *opt.Interactive {
		opt.Log("\n\033[32mBuild successful\033[0m\n")
		if m["serverless"] == "true" {
			opt.Log("\033[90m" + "The compiled serverless function has been saved in the .zeabur directory." + "\033[0m")
		} else {
			opt.Log("\033[90m" + "To run the image, use the following command:" + "\033[0m")
			if m["outputDir"] != "" || (t == types.PlanTypeStatic && m["serverless"] == "true") {
				opt.Log("npx serve .zeabur/output/static")
			} else {
				opt.Log("docker run -p 8080:8080 -e PORT=8080 -it " + *opt.ResultImage)
			}
		}
	}

	return nil
}

func (opt *BuildOptions) Log(msg string, args ...any) {
	_, _ = fmt.Fprintf(opt.LogWriter, msg, args...)
}
