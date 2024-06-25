package elixir

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

var elixirVersions = map[string]string{
	"1.7":  "1.7",
	"1.8":  "1.8",
	"1.9":  "1.9",
	"1.10": "1.10",
	"1.11": "1.11",
	"1.12": "1.12",
	"1.13": "1.13",
	"1.14": "1.14",
	"1.15": "1.15",
}

// DetermineElixirVersion returns the version of Elixir.
func DetermineElixirVersion(src afero.Fs) (string, error) {
	fileName := "mix.exs"
	if utils.HasFile(src, fileName) {
		content, err := utils.ReadFileToUTF8(src, fileName)
		if err != nil {
			return "", err
		}

		// Format: ``` elixir: "~> 1.12", ```
		pattern := regexp.MustCompile(`elixir: "~> ([0-9.]+)"`)
		// Search for the target version in the file.
		matches := pattern.FindStringSubmatch(string(content))
		if len(matches) > 0 {
			if version := elixirVersions[matches[1]]; version != "" {
				return version, nil
			}
			// For future versions use latest
			return "latest", nil
		}

		return "", errors.New("unable to determine Elixir version")
	}

	return "", errors.New("unable to determine Elixir version")
}

// DetermineElixirFramework returns the framework being used (e.g Phoenix).
func DetermineElixirFramework(src afero.Fs) (string, error) {
	filename := "mix.exs"
	if utils.HasFile(src, filename) {
		content, err := utils.ReadFileToUTF8(src, filename)
		if err != nil {
			return "", err
		}

		// Detect which framework is being used
		if utils.WeakContains(string(content), ":phoenix") {
			return string(types.ElixirFrameworkPhoenix), nil
		}

		// if no framework is being used return empty
		return "", nil
	}

	return "", errors.New("unable to determine Elixir framework")
}

// CheckElixirEcto returns true if Elixir is using Ecto.
func CheckElixirEcto(src afero.Fs) (string, error) {
	fileName := "mix.exs"
	if utils.HasFile(src, fileName) {
		content, err := utils.ReadFileToUTF8(src, fileName)
		if err != nil {
			return "", err
		}

		ectoFound := utils.WeakContains(string(content), "ecto_sql")
		postgrexFound := utils.WeakContains(string(content), "postgrex")

		usesEcto := strconv.FormatBool(ectoFound && postgrexFound)
		return usesEcto, nil
	}

	return "", errors.New("unable to determine if Ecto is used")
}
