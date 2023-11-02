package zeaburpack

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/pan93412/envexpander"
)

type buildImageOptions struct {
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

		dockerfileEnv += "ENV " + key + " " + value + "\n"
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

	dockerCmd := []string{
		"buildx",
		"build",
		"-t", opt.ResultImage,
		"-f", dockerfilePath,
	}

	if opt.PlainDockerProgress {
		dockerCmd = append(dockerCmd, "--progress", "plain")
	} else {
		dockerCmd = append(dockerCmd, "--progress", "tty")
	}

	if opt.CacheFrom != nil && len(*opt.CacheFrom) > 0 {
		dockerCmd = append(dockerCmd, "--cache-from", *opt.CacheFrom)
	}

	if opt.CacheTo != nil && len(*opt.CacheTo) > 0 {
		dockerCmd = append(dockerCmd, "--cache-to", *opt.CacheTo)
	}

	dockerCmd = append(dockerCmd, opt.AbsPath)

	cmd := exec.Command("docker", dockerCmd...)
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

	if opt.HandleLog == nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			println("failed to run docker build: " + err.Error())
			return fmt.Errorf("run docker build: %w", err)
		}
		return nil
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("get stderr pipe: %w", err)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start docker build: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(errPipe)
		for scanner.Scan() {
			t := scanner.Text()
			println(t)
			(*opt.HandleLog)(t)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			(*opt.HandleLog)(scanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("wait docker build: %w", err)
	}

	return nil
}
