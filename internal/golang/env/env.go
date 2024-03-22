package golangenv

import (
	"log/slog"
	"os/exec"
	"regexp"

	"github.com/zeabur/action/environment"
)

func init() {
	environment.RegisterSoftware(&Golang{})
}

// Golang represents the Golang SDK.
type Golang struct{}

func (g Golang) Name() string {
	return "go"
}

func (g Golang) Version() (string, bool) {
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err == nil {
		// go version go1.22.0 darwin/arm64
		goVersionExtractor := regexp.MustCompile(`go([\d.]+)`)
		matches := goVersionExtractor.FindSubmatch(out)

		if len(matches) > 1 {
			return string(matches[1]), true
		}

		slog.Debug("failed to extract Golang version", "output", string(out))
	} else {
		slog.Debug("failed to get Python version", "error", err)
	}

	return "", false
}
