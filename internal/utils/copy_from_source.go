package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CopyFromSource copies a directory from source code to the host
func CopyFromSource(dirInSrc, destOnHost string) error {
	if err := os.MkdirAll(".tmp", 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	defer func() {
		removeCmd := exec.Command("rm", "-rf", ".tmp")
		removeCmd.Stderr = os.Stderr
		if err := removeCmd.Run(); err != nil {
			fmt.Println(err)
		}
	}()
	var stderr strings.Builder
	tempCopyCmd := exec.Command("cp", "-r", dirInSrc, ".tmp")
	tempCopyCmd.Stderr = &stderr
	err := tempCopyCmd.Run()
	if err != nil {
		return fmt.Errorf("copy from source code: %s: %w", stderr.String(), err)
	}

	if err := os.MkdirAll(destOnHost, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	copyCmd := exec.Command("cp", "-r", ".tmp/.", destOnHost)

	copyCmd.Stderr = &stderr
	err = copyCmd.Run()
	if err != nil {
		return fmt.Errorf("copy from source code: %s: %w", stderr.String(), err)
	}

	return nil
}
