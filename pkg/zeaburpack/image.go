package zeaburpack

import (
	"bufio"
	"fmt"
	"io"
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
	srcDIr := opt.AbsPath
	buildID := strconv.Itoa(rand.Int())

	err := os.MkdirAll(path.Join(tempDir, buildID), 0o755)
	if err != nil {
		return err
	}

	dockerfilePath := path.Join(tempDir, buildID, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(newDockerfile), 0o644)
	if err != nil {
		return fmt.Errorf("write Dockerfile: %w", err)
	}

	dockerIgnore := []string{".next", "node_modules", ".zeabur"}
	dockerIgnorePath := path.Join(srcDIr, ".dockerignore")
	isDockerIgnoreExists := false
	// if .dockerignore exists, we need to append the content to the end of the file
	if fileExists(dockerIgnorePath) {
		isDockerIgnoreExists = true
		err := backupFile(dockerIgnorePath, dockerIgnorePath+".bak")
		if err != nil {
			return fmt.Errorf("copy .dockerignore: %w", err)
		}
		dockerIgnoreFile, err := os.ReadFile(path.Join(srcDIr, ".dockerignore"))
		if err != nil {
			return fmt.Errorf("read .dockerignore: %w", err)
		}
		dockerIgnore = append(dockerIgnore, strings.Split(string(dockerIgnoreFile), "\n")...)
		err = appendToFile(dockerIgnorePath, strings.Join(dockerIgnore, "\n"))
		if err != nil {
			return fmt.Errorf("append .dockerignore: %w", err)
		}
	}
	// if .dockerignore does not exist, we need to create it
	err = os.WriteFile(dockerIgnorePath, []byte(strings.Join(dockerIgnore, "\n")), 0o644)
	if err != nil {
		return fmt.Errorf("write .dockerignore: %w", err)
	}

	defer func() {
		if !isDockerIgnoreExists {
			// remove the .dockerignore file we created
			err := os.Remove(dockerIgnorePath)
			if err != nil {
				println("failed to remove .dockerignore: " + err.Error())
			}
		} else {
			// restore the .dockerignore file
			err = os.Rename(dockerIgnorePath+".bak", dockerIgnorePath)
			if err != nil {
				println("failed to restore .dockerignore: " + err.Error())
			}
		}
	}()

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
			return err
		}
		return nil
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(errPipe)
		for scanner.Scan() {
			(*opt.HandleLog)(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			(*opt.HandleLog)(scanner.Text())
		}
	}()

	return cmd.Wait()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func appendToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func backupFile(sourceFile, destinationFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return nil
}
