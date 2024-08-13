package utils

import (
	"regexp"
	"strings"
)

// ExtractErlangEntryFromGleamEntrypointShell extracts the Erlang entrypoint from the Gleam entrypoint shell script
func ExtractErlangEntryFromGleamEntrypointShell(gleamEntryShell string) string {
	packageRe := regexp.MustCompile(`PACKAGE=([^\s]+)`)
	packageMatch := packageRe.FindStringSubmatch(gleamEntryShell)
	if len(packageMatch) > 1 {
		packageName := packageMatch[1]

		evalRe := regexp.MustCompile(`-eval\s+"([^"]+)"`)
		evalMatch := evalRe.FindStringSubmatch(gleamEntryShell)
		if len(evalMatch) > 1 {
			evalString := evalMatch[1]

			result := strings.ReplaceAll(evalString, "$PACKAGE", packageName)
			return result
		}
	}

	return ""
}
