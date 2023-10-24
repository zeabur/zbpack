package dotnet

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// DetermineFramework is used to determine the Dotnet framework
func DetermineFramework(entryPoint string, src afero.Fs) (string, error) {
	fileName := entryPoint + ".csproj"
	content, err := afero.ReadFile(src, fileName)
	if err == nil {
		pattern := regexp.MustCompile(`Project Sdk="(.*?)"`)
		// Search for the target framework in the file.
		matches := pattern.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			if strings.Contains(matches[1], "Microsoft.NET.Sdk.BlazorWebAssembly") {
				return string(types.DotnetFrameworkBlazorWasm), nil
			} else if strings.Contains(matches[1], "Microsoft.NET.Sdk.Web") {
				return string(types.DotnetFrameworkAspnet), nil
			} else if strings.Contains(matches[1], "Microsoft.NET.Sdk") {
				return string(types.DotnetFrameworkConsole), nil
			}
		}
	}

	return "", errors.New("Unable to determine framework")
}

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
