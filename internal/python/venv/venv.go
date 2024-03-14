// Package venv is the utils for zbpack/python/prepare.
package venv

import (
	"fmt"
	"maps"
	"path/filepath"
	"strings"
	"sync"

	"log/slog"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/pkg/types"
)

// VirtualEnvironmentContext is the context of a virtual environment.
type VirtualEnvironmentContext struct {
	PackageManager types.PythonPackageManager
	Path           string
	PathGetter     func() (string, error)
}

// GetPath returns the Path of the virtual environment.
func (v VirtualEnvironmentContext) GetPath() (string, error) {
	if v.Path != "" {
		return v.Path, nil
	}

	if v.PathGetter == nil {
		return "", fmt.Errorf("no path are given")
	}

	path, err := v.PathGetter()
	if err != nil {
		return "", fmt.Errorf("calling PathGetter(): %w", err)
	}
	v.Path = strings.TrimSpace(path)
	return v.Path, nil
}

// GetSitePackagesDirectory returns the site-packages directory of the virtual environment.
func (v VirtualEnvironmentContext) GetSitePackagesDirectory() (string, error) {
	path, err := v.GetPath()
	if err != nil {
		return "", fmt.Errorf("get path: %w", err)
	}

	sitePackagesPath, err := filepath.Glob(filepath.Join(path, "lib/*/site-packages"))
	if err != nil || len(sitePackagesPath) == 0 {
		return "", fmt.Errorf("no site-packages in this venv: %s", path)
	}

	return sitePackagesPath[0], nil
}

// GetPackageManager returns the PackageManager of the virtual environment.
func (v VirtualEnvironmentContext) GetPackageManager() types.PythonPackageManager {
	return v.PackageManager
}

// PutEnv puts the virtual environment into the environment variables.
//
// It acts like `bin/activate` command.
func (v VirtualEnvironmentContext) PutEnv(currentEnv zbaction.EnvironmentVariables) zbaction.EnvironmentVariables {
	if v.GetPackageManager() != types.PythonPackageManagerPip {
		// Only the vanilla PIP requires such a hack.
		// For other package managers, using their built-in commands is enough.
		return currentEnv
	}

	path, err := v.GetPath()
	if err != nil {
		slog.Warn("failed to get the path of the virtual environment – fallback", slog.String("error", err.Error()))
		return currentEnv
	}

	newEnv := maps.Clone(currentEnv)

	// VIRTUAL_ENV will be read by cpython/launcher.c and used to set sys.prefix.
	newEnv["VIRTUAL_ENV"] = path

	if oldPath, ok := newEnv["PATH"]; ok {
		newEnv["PATH"] = filepath.Join(path, "bin") + ":" + oldPath
	} else {
		newEnv["PATH"] = filepath.Join(path, "bin")
	}

	return newEnv
}

type globalPreparedRegistry struct {
	venvContext map[zbaction.JobID]*VirtualEnvironmentContext

	lock *sync.Mutex
}

var registry = globalPreparedRegistry{
	venvContext: make(map[zbaction.JobID]*VirtualEnvironmentContext),
	lock:        &sync.Mutex{},
}

// RegisterVenvContext registers a virtual environment context.
//
// It should be only used by zbpack/python/prepare.
func RegisterVenvContext(jobID zbaction.JobID, venv *VirtualEnvironmentContext) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	registry.venvContext[jobID] = venv
}

// DropVenvContext drops a virtual environment context.
//
// It should be only used by zbpack/python/prepare.
func DropVenvContext(jobID zbaction.JobID) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	delete(registry.venvContext, jobID)
}

// GetVenvContext gets a virtual environment context.
//
// You should call it after running zbpack/python/prepare.
func GetVenvContext(jobID zbaction.JobID) (*VirtualEnvironmentContext, error) {
	registry.lock.Lock()
	defer registry.lock.Unlock()

	venv, ok := registry.venvContext[jobID]
	if !ok {
		return nil, fmt.Errorf("zbpack/python/prepare did not run in %s", jobID)
	}

	return venv, nil
}
