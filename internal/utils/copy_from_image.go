package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
	return nil
}
