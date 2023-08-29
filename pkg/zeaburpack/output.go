package zeaburpack

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// copyZeaburOutputToHost copies the .zeabur/output directory from the result image to the host
func copyZeaburOutputToHost(resultImage, targetDir string) (bool, error) {
	createCmd := exec.Command("docker", "create", resultImage)
	createCmd.Stderr = os.Stderr
	output, err := createCmd.Output()
	if err != nil {
		return false, err
	}

	defer func() {
		removeCmd := exec.Command("docker", "rm", "-f", strings.TrimSpace(string(output)))
		removeCmd.Stderr = os.Stderr
		if err := removeCmd.Run(); err != nil {
			log.Println(err)
		}
	}()

	containerID := strings.TrimSpace(string(output))

	if err := os.RemoveAll(path.Join(targetDir, ".zeabur")); err != nil {
		log.Printf("failed to delete .zeabur directory: %s", err)
	}

	if err := os.MkdirAll(path.Join(targetDir, ".zeabur"), 0o755); err != nil {
		log.Printf("failed to create .zeabur directory: %s", err)
	}

	copyCmd := exec.Command("docker", "cp", containerID+":/src/.zeabur/output/.", path.Join(targetDir, ".zeabur/output"))
	var stderr strings.Builder
	copyCmd.Stderr = &stderr
	err = copyCmd.Run()
	if err != nil {
		if strings.Contains(stderr.String(), "Could not find the file /src/.zeabur/output") {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return true, nil
}
