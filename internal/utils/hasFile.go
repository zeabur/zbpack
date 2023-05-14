package utils

import "github.com/spf13/afero"

// HasFile checks if the given file exists in the given filesystem.
func HasFile(src afero.Fs, fileNames ...string) bool {
	for _, fileName := range fileNames {
		if exists, _ := afero.Exists(src, fileName); exists {
			return true
		}
	}
	return false
}
