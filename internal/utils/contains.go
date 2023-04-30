package utils

import "strings"

// Contains is just like strings.Contains, but it's case-insensitive.
func Contains(a, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b))
}
