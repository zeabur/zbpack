package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CopyFromImage copies a directory from a docker image to the host
func CopyFromImage(image, srcInImage, destOnHost string) error {
	createCmd := exec.Command("docker", "create", image)
	createCmd.Stderr = os.Stderr
	output, err := createCmd.Output()
	if err != nil {
		return fmt.Errorf("create docker container: %w", err)
	}

	defer func() {
		removeCmd := exec.Command("docker", "rm", "-f", strings.TrimSpace(string(output)))
		removeCmd.Stderr = os.Stderr
		if err := removeCmd.Run(); err != nil {
			log.Println(err)
		}
	}()

	containerID := strings.TrimSpace(string(output))

	if err := os.MkdirAll(destOnHost, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	copyCmd := exec.Command("docker", "cp", containerID+":"+srcInImage, destOnHost)
	var stderr strings.Builder
	copyCmd.Stderr = &stderr
	err = copyCmd.Run()
	if err != nil {
		return fmt.Errorf("copy from image: %s: %w", stderr.String(), err)
	}

	// there may be some symlinks in the copied directory, we need to resolve them
	err = filepath.Walk(destOnHost, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			targetInImage, err := os.Readlink(p)
			if err != nil {
				return fmt.Errorf("read symlink target: %w", err)
			}

			relativeToSrc, err := filepath.Rel(srcInImage, targetInImage)
			if err != nil {
				return fmt.Errorf("calculate relative path: %w", err)
			}

			absTarget := filepath.Join(destOnHost, relativeToSrc)
			err = os.Remove(p)
			if err != nil {
				return fmt.Errorf("remove symlink: %w", err)
			}

			err = os.Symlink(absTarget, p)
			if err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
		}

		return nil
	})

	return nil
}
