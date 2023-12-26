package utils

import (
	"io"
	"os"
	"path/filepath"
)

// Copy copies a file or directory from src to dst.
func Copy(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return _copy(src, dst, info)
}

func _copy(src, dst string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}
	if info.IsDir() {
		return copyDirectory(src, dst, info)
	}
	return copyFile(src, dst, info)
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(src), target)
	}
	return Copy(target, dst)
}

func copyDirectory(src, dst string, info os.FileInfo) error {
	err := os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())
		if err := Copy(srcFile, dstFile); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
