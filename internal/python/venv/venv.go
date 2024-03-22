// Package venv is the utils for zbpack/python/prepare.
package venv

import (
	"fmt"
	"maps"
	"path/filepath"
	"sync"

	zbaction "github.com/zeabur/action"
)

// VirtualEnvironmentContext is the context of a virtual environment.
type VirtualEnvironmentContext struct {
	Path string
}

// GetPath returns the Path of the virtual environment.
func (v VirtualEnvironmentContext) GetPath() string {
	return v.Path
}

// GetSitePackagesDirectory returns the site-packages directory of the virtual environment.
func (v VirtualEnvironmentContext) GetSitePackagesDirectory() (string, error) {
	path := v.GetPath()

	sitePackagesPath, err := filepath.Glob(filepath.Join(path, "lib/*/site-packages"))
	if err != nil || len(sitePackagesPath) == 0 {
		return "", fmt.Errorf("no site-packages in this venv: %s", path)
	}

	return sitePackagesPath[0], nil
}

// PutEnv puts the virtual environment into the environment variables.
//
// It acts like `bin/activate` command.
func (v VirtualEnvironmentContext) PutEnv(currentEnv zbaction.EnvironmentVariables) zbaction.EnvironmentVariables {
	path := v.GetPath()
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
