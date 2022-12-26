package utils

import "os"

func HasFile(dirPath string, fileNames ...string) bool {
	for _, fileName := range fileNames {
		if _, err := os.Stat(dirPath + "/" + fileName); err == nil {
			return true
		}
	}
	return false
}
