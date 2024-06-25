package utils

import (
	"fmt"

	"github.com/spf13/afero"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// ReadToUTF8 accepts a byte slice and check if it is a UTF-8, UTF-16 LE
// or UTF-16 BE encoded string by BOM and decode it to UTF-8 byte string.
func ReadToUTF8(content []byte) ([]byte, error) {
	// If this content does not contain BOM, we assume it is UTF-8.
	fallbackEncoder := unicode.UTF8.NewDecoder()

	// Use unicode.BOMOverride to check the encoding of the content.
	// If the content is not UTF-8, it will be decoded to UTF-8.
	decoder := unicode.BOMOverride(fallbackEncoder)

	// Decode the content to UTF-8.
	content, _, err := transform.Bytes(decoder, content)
	if err != nil {
		return nil, fmt.Errorf("unexpected encoding: %w", err)
	}

	return content, nil
}

// ReadFileToUTF8 reads a file from the filesystem and decode it to UTF-8.
//
// It is basically the wrapper of afero.ReadFile.
func ReadFileToUTF8(fs afero.Fs, path string) ([]byte, error) {
	content, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	decoded, err := ReadToUTF8(content)
	if err != nil {
		return nil, fmt.Errorf("decode file %s: %w", path, err)
	}

	return decoded, nil
}
