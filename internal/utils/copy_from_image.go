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

	err = deleteHiddenFilesAndDirs(destOnHost)
	if err != nil {
		return fmt.Errorf("delete hidden files and directories in directory: %w", err)
	}

	return nil
}

// DeleteHiddenFilesAndDirs deletes hidden files and directories in a directory
func deleteHiddenFilesAndDirs(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	entries, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			entryPath := filepath.Join(dirPath, entry.Name())

			if entry.IsDir() {
				if err := os.RemoveAll(entryPath); err != nil {
					return err
				}
			} else {
				if err := os.Remove(entryPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
