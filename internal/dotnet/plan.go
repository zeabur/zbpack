package dotnet

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
)

// DetermineSDKVersion returns the version of the SDK.
func DetermineSDKVersion(entryPoint string, src afero.Fs) (string, error) {
	fileName := entryPoint + ".csproj"
	if utils.HasFile(src, fileName) {
		content, err := afero.ReadFile(src, fileName)
		if err != nil {
			return "", err
		}

		pattern := regexp.MustCompile(`<TargetFramework>([^<]+)</TargetFramework>`)
		// Search for the target framework in the file.
		matches := pattern.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return strings.Replace(matches[1], "net", "", -1), nil
		}
	}

	return "", errors.New("Unable to determine SDK version")
}
