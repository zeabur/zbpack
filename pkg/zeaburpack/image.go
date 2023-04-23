package zeaburpack

import (
	"bufio"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type BuildImageOptions struct {
	Dockerfile          string
	AbsPath             string
	UserVars            map[string]string
	ResultImage         string
	HandleLog           *func(log string)
	PlainDockerProgress bool
	CacheFrom           *string
}

func buildImage(opt *BuildImageOptions) error {

	lines := strings.Split(opt.Dockerfile, "\n")
	stageLines := []int{}
	for i, line := range lines {
		if strings.HasPrefix(line, "FROM") {
			stageLines = append(stageLines, i)
		}
	}

	dockerfileEnv := ""
	for key, value := range opt.UserVars {
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

	err := os.MkdirAll(path.Join(tempDir, buildID), 0755)
	if err != nil {
		return err
	}

	dockerfilePath := path.Join(tempDir, buildID, "Dockerfile")
	if err := os.WriteFile(
		dockerfilePath, []byte(newDockerfile), 0644,
	); err != nil {
		return err
	}

	dockerCmd := []string{
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
		// if cacheFrom contains tag, we need to remove it
		if strings.Contains(*opt.CacheFrom, ":") {
			*opt.CacheFrom = strings.Split(*opt.CacheFrom, ":")[0]
		}
		dockerCmd = append(dockerCmd, "--cache-from", *opt.CacheFrom)
		dockerCmd = append(dockerCmd, "--build-arg", "BUILDKIT_INLINE_CACHE=1")
	}

	dockerCmd = append(dockerCmd, opt.AbsPath)

	println("DOCKER_BUILDKIT=1 docker " + strings.Join(dockerCmd, " "))
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

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
