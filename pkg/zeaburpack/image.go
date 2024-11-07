package zeaburpack

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/zeabur/zbpack/pkg/types"
)

type buildImageOptions struct {
	PlanType            types.PlanType
	PlanMeta            types.PlanMeta
	Dockerfile          string
	AbsPath             string
	UserVars            map[string]string
	ResultImage         string
	PlainDockerProgress bool

	CacheFrom *string
	CacheTo   *string

	// PushImage is a flag to indicate if the image should be pushed to the registry.
	PushImage bool

	// LogWriter is a [io.Writer] that will be written when a log is emitted.
	// nil to use the default log writer.
	LogWriter io.Writer
}

// ServerlessTarPath is the path to the serverless output tar file
var ServerlessTarPath = filepath.Join(
	lo.Must(os.MkdirTemp("", "zbpack-buildkit-artifact-*")),
	"serverless-output.tar",
)

func buildImage(opt *buildImageOptions) error {
	if opt.LogWriter == nil {
		opt.LogWriter = os.Stderr
	}

	tempDir := os.TempDir()
	buildID := strconv.Itoa(rand.Int())

	err := os.MkdirAll(path.Join(tempDir, buildID), 0o755)
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	dockerfilePath := path.Join(tempDir, buildID, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(opt.Dockerfile), 0o644)
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

	if opt.PlanMeta["serverless"] == "true" || opt.PlanType == types.PlanTypeNix {
		buildKitCmd = append(buildKitCmd, "--output", "type=tar,dest="+ServerlessTarPath)
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
		buildKitCmd = append(buildKitCmd, "--import-cache", "type=registry,ref="+*opt.CacheFrom)
	}

	if opt.CacheTo != nil && len(*opt.CacheTo) > 0 {
		buildKitCmd = append(buildKitCmd, "--export-cache", "type=registry,ref="+*opt.CacheTo)
	}

	if opt.PlainDockerProgress {
		buildKitCmd = append(buildKitCmd, "--progress", "plain")
	} else {
		buildKitCmd = append(buildKitCmd, "--progress", "tty")
	}

	buildctlCmd := exec.Command("buildctl", buildKitCmd...)
	buildctlCmd.Stderr = opt.LogWriter
	output, err := buildctlCmd.Output()
	if err != nil {
		return fmt.Errorf("run buildctl build: %w", err)
	}

	if !strings.Contains(buildctlCmd.String(), "type=docker") {
		return nil // buildctl have handled push
	}

	dockerLoadCmd := exec.Command("docker", "load")
	dockerLoadCmd.Stdin = bytes.NewReader(output)
	dockerLoadCmd.Stdout = opt.LogWriter
	dockerLoadCmd.Stderr = opt.LogWriter
	if err := dockerLoadCmd.Run(); err != nil {
		return fmt.Errorf("run docker load: %w", err)
	}

	return nil
}
