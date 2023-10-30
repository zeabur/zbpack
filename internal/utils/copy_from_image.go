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
	excludeFiles := []string{".gitkeep", ".ini", ".env"}
	excludeDirs := []string{".git"}
	err = deleteFilesInDirectory(excludeFiles, destOnHost)
	if err != nil {
		return fmt.Errorf("delete files in directory: %w", err)
	}
	err = deleteDirectoriesInDirectory(excludeDirs, destOnHost)
	if err != nil {
		return fmt.Errorf("delete directories in directory: %w", err)
	}

	return nil
}

func deleteFilesInDirectory(deleteFiles []string, path string) error {
	// walk through the directory
	err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			fileName := fileInfo.Name()

			for _, targetFile := range deleteFiles {
				if fileName == targetFile {
					filePath := filepath.Join(path, fileName)

					err := os.Remove(filePath)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func deleteDirectoriesInDirectory(deleteDirs []string, path string) error {
	fileInfo, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, dirInfo := range fileInfo {
		if dirInfo.IsDir() {
			dirName := dirInfo.Name()

			for _, targetDir := range deleteDirs {
				if dirName == targetDir {
					dirPath := filepath.Join(path, dirName)
					err := os.RemoveAll(dirPath)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
