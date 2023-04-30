package utils

import "strings"

// WeakContains is just like strings.Contains, but it's case-insensitive.
func WeakContains(a, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b))
}
