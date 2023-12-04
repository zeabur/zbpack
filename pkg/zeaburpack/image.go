package zeaburpack

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/pan93412/envexpander"
	"github.com/zeabur/zbpack/pkg/types"
)

type buildImageOptions struct {
	PlanType            types.PlanType
	PlanMeta            types.PlanMeta
	Dockerfile          string
	AbsPath             string
	UserVars            map[string]string
	ResultImage         string
	HandleLog           *func(log string)
	PlainDockerProgress bool

	CacheFrom *string
	CacheTo   *string

	// ProxyRegistry is the registry to be used for the image.
	// See referenceConstructor for more details.
	ProxyRegistry *string

	// PushImage is a flag to indicate if the image should be pushed to the registry.
	PushImage bool
}

func buildImage(opt *buildImageOptions) error {
	// resolve env variable statically and don't depend on Dockerfile's order
	resolvedVars := envexpander.ResolveEnvVariable(opt.UserVars)

	refConstructor := newReferenceConstructor(opt.ProxyRegistry)
	lines := strings.Split(opt.Dockerfile, "\n")
	stageLines := make([]int, 0)

	for i, line := range lines {
		fromStatement, isFromStatement := ParseFrom(line)
		if !isFromStatement {
			continue
		}

		// Construct the reference.
		newRef := refConstructor.Construct(fromStatement.Source)

		// Replace this FROM line.
		fromStatement.Source = newRef
		lines[i] = fromStatement.String()

		// Mark this FROM line as a stage.
		if stage, ok := fromStatement.Stage.Get(); ok {
			refConstructor.AddStage(stage)
		}
		stageLines = append(stageLines, i)
	}

	// sort the resolvedVars by key so we can build
	// the reproducible dockerfile
	sortedResolvedVarsKey := make([]string, 0, len(resolvedVars))
	for key := range resolvedVars {
		sortedResolvedVarsKey = append(sortedResolvedVarsKey, key)
	}
	sort.Strings(sortedResolvedVarsKey)

	// build the dockerfile
	dockerfileEnv := ""

	// Inject CI env so everyone knows that we are a CI.
	if _, ok := resolvedVars["CI"]; !ok {
		dockerfileEnv += "ENV CI true\n"
	}

	for _, key := range sortedResolvedVarsKey {
		value := resolvedVars[key]

		// skip empty value
		if len(value) == 0 {
			continue
		}

		value = strings.ReplaceAll(value, "\n", "\\n")
		value = strings.ReplaceAll(value, "'", "\\'")
		value = strings.ReplaceAll(value, "\"", "\\\"")
		value = strings.ReplaceAll(value, "\\", "\\\\")

		dockerfileEnv += "ENV " + key + " \"" + value + "\"\n"
	}

	for _, stageLine := range stageLines {
		lines[stageLine] = lines[stageLine] + "\n" + dockerfileEnv + "\n"
	}
	newDockerfile := strings.Join(lines, "\n")

	tempDir := os.TempDir()
	buildID := strconv.Itoa(rand.Int())

	err := os.MkdirAll(path.Join(tempDir, buildID), 0o755)
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	dockerfilePath := path.Join(tempDir, buildID, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(newDockerfile), 0o644)
	if err != nil {
		return fmt.Errorf("write Dockerfile: %w", err)
	}

	dockerIgnore := []string{".next", "node_modules", ".zeabur"}
	dockerIgnorePath := path.Join(tempDir, buildID, ".dockerignore")
	err = os.WriteFile(dockerIgnorePath, []byte(strings.Join(dockerIgnore, "\n")), 0o644)
	if err != nil {
		return fmt.Errorf("write .dockerignore: %w", err)
	}

	buildKitCmd := []string{
		"build",
		"--frontend", "dockerfile.v0",
		"--local", "context=" + opt.AbsPath,
		"--local", "dockerfile=" + path.Dir(dockerfilePath),
	}

	if opt.PlanMeta["serverless"] == "true" || opt.PlanMeta["outputDir"] != "" || opt.PlanType == types.PlanTypeStatic {
		buildKitCmd = append(buildKitCmd, "--output", "type=local,dest="+path.Join(os.TempDir(), "zbpack/buildkit"))
	} else {
		t := "image"
		if !opt.PushImage {
			// -> docker registry
			t = "docker"
		}

		o := "type=" + t + ",name=" + opt.ResultImage
		if opt.PushImage {
			o += ",push=true"
		}
		buildKitCmd = append(buildKitCmd, "--output", o)
	}

	if opt.CacheFrom != nil && len(*opt.CacheFrom) > 0 {
		buildKitCmd = append(buildKitCmd, "--import-cache type=registry,ref="+*opt.CacheFrom)
	}

	if opt.CacheTo != nil && len(*opt.CacheTo) > 0 {
		buildKitCmd = append(buildKitCmd, "--export-cache", *opt.CacheTo)
	}

	if opt.PlainDockerProgress {
		buildKitCmd = append(buildKitCmd, "--progress", "plain")
	} else {
		buildKitCmd = append(buildKitCmd, "--progress", "tty")
	}

	buildctlCmd := exec.Command("buildctl", buildKitCmd...)
	buildctlCmd.Stderr = NewHandledWriter(os.Stderr, opt.HandleLog)
	output, err := buildctlCmd.Output()
	if err != nil {
		return fmt.Errorf("run buildctl build: %w", err)
	}

	if opt.PushImage {
		return nil // buildctl have handled push
	}

	dockerLoadCmd := exec.Command("docker", "load")
	dockerLoadCmd.Stdin = bytes.NewReader(output)
	dockerLoadCmd.Stdout = NewHandledWriter(os.Stdout, opt.HandleLog)
	dockerLoadCmd.Stderr = NewHandledWriter(os.Stderr, opt.HandleLog)
	if err := dockerLoadCmd.Run(); err != nil {
		return fmt.Errorf("run docker load: %w", err)
	}

	return nil
}
