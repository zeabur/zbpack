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
	Dockerfile  string
	AbsPath     string
	UserVars    map[string]string
	ResultImage string
	HandleLog   func(log string)
}

func buildImage(opt *BuildImageOptions) error {

	lines := strings.Split(opt.Dockerfile, "\n")
	firstLine := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "FROM") {
			firstLine = i
			break
		}
	}

	dockerfileEnv := ""
	for key, value := range opt.UserVars {
		dockerfileEnv += "ENV " + key + " " + value + "\n"
	}

	lines[firstLine] = lines[firstLine] + "\n" + dockerfileEnv + "\n"
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

	cmd := exec.Command(
		"docker", "build",
		"-t", opt.ResultImage,
		"-f", dockerfilePath,
		"--progress", "plain",
		opt.AbsPath,
	)

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
			opt.HandleLog(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			opt.HandleLog(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
