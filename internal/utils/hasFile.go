package utils

import (
	"github.com/zeabur/zbpack/internal/source"
)

func HasFile(src *source.Source, fileNames ...string) bool {
	for _, fileName := range fileNames {
		if (*src).HasFile(fileName) {
			return true
		}
	}
	return false
}
