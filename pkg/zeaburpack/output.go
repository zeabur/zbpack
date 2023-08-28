package zeaburpack

import (
	"fmt"
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
		removeCmd.Run()
	}()

	containerID := strings.TrimSpace(string(output))

	dir := path.Join(targetDir, ".zeabur")
	copyCmd := exec.Command("docker", "cp", containerID+":/src/.zeabur/output", dir)
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
