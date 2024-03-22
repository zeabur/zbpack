// Package pythonenv provides the Python environment checkers.
package pythonenv

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/zeabur/action/environment"
)

func init() {
	environment.RegisterSoftware(&Python{})
	environment.RegisterSoftware(&Pip{})
	environment.RegisterSoftware(&Pipenv{})
	environment.RegisterSoftware(&Poetry{})
	environment.RegisterSoftware(&Pdm{})
	environment.RegisterSoftware(&Rye{})
}

// Python represents the Python SDK.
type Python struct{}

// Name returns the name of the Python SDK.
func (p Python) Name() string {
	return "python"
}

// Version returns the version of the Python SDK.
func (p Python) Version() (string, bool) {
	exeCandidates := []string{"python3", "python"}
	var exe string
	for _, e := range exeCandidates {
		if _, err := exec.LookPath(e); err == nil {
			exe = e
			break
		}
	}
	if exe == "" {
		slog.Debug("Python not found", slog.Any("candidates", exeCandidates))
		return "", false
	}

	cmd := exec.Command(exe, "--version")
	out, err := cmd.Output()
	if err == nil {
		// Python 3.11.7
		version, ok := strings.CutPrefix(string(out), "Python ")

		if ok {
			return strings.TrimSpace(version), true
		}
		slog.Debug("failed to extract Python version", "output", string(out))
	} else {
		slog.Debug("failed to get Python version", "error", err)
	}

	return "", false
}

// Pip represents the Pip package manager (actually, `uv`).
type Pip struct{}

// Name returns the name of the Pip (Uv) package manager.
func (p Pip) Name() string {
	return "pip"
}

// Version returns the version of the Pip (Uv) package manager.
func (p Pip) Version() (string, bool) {
	cmd := exec.Command("uv", "-V")
	out, err := cmd.Output()
	if err == nil {
		// uv 0.1.14
		version, ok := strings.CutPrefix(string(out), "uv ")

		if ok {
			return strings.TrimSpace(version), true
		}
		slog.Debug("failed to extract Pip version", "output", string(out))
	} else {
		slog.Debug("failed to get Pip version", "error", err)
	}

	return "", false
}

// Pipenv represents the pipenv package manager.
type Pipenv struct{}

// Name returns the name of the Pipenv package manager.
func (p Pipenv) Name() string {
	return "pipenv"
}

// Version returns the version of the Pipenv package manager.
func (p Pipenv) Version() (string, bool) {
	cmd := exec.Command("pipenv", "--version")
	out, err := cmd.Output()
	if err == nil {
		// pipenv, version 2021.5.29
		version, ok := strings.CutPrefix(string(out), "pipenv, version ")

		if ok {
			return strings.TrimSpace(version), true
		}
		slog.Debug("failed to extract Pipenv version", "output", string(out))
	} else {
		slog.Debug("failed to get Pipenv version", "error", err)
	}

	return "", false
}

// Poetry represents the poetry package manager.
type Poetry struct{}

// Name returns the name of the Poetry package manager.
func (p Poetry) Name() string {
	return "poetry"
}

// Version returns the version of the Poetry package manager.
func (p Poetry) Version() (string, bool) {
	cmd := exec.Command("poetry", "-V", "--no-ansi")
	out, err := cmd.Output()
	if err == nil {
		// Poetry (version 1.7.1)
		_, versionNum, ok := strings.Cut(string(out), "version ")
		if ok {
			return strings.TrimRight(strings.TrimSpace(versionNum), ")"), true
		}
		slog.Debug("failed to extract Poetry version", "output", string(out))
	} else {
		slog.Debug("failed to get Poetry version", "error", err)
	}

	return "", false
}

// Pdm represents the pdm package manager.
type Pdm struct{}

// Name returns the name of the Pdm package manager.
func (p Pdm) Name() string {
	return "pdm"
}

// Version returns the version of the Pdm package manager.
func (p Pdm) Version() (string, bool) {
	cmd := exec.Command("pdm", "-V")
	_, err := cmd.Output()
	if err == nil {
		// FIXME: implement version
		return "", true
	} else {
		slog.Debug("failed to get PDM version", "error", err)
	}

	return "", false
}

// Rye represents the rye package manager.
type Rye struct{}

// Name returns the name of the Rye package manager.
func (r Rye) Name() string {
	return "rye"
}

// Version returns the version of the Rye package manager.
func (r Rye) Version() (string, bool) {
	cmd := exec.Command("rye", "--version")
	_, err := cmd.Output()
	if err == nil {
		// FIXME: implement version
		return "", true
	} else {
		slog.Debug("failed to get Rye version", "error", err)
	}

	return "", false
}
