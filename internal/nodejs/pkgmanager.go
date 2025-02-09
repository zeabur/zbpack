package nodejs

import (
	"fmt"

	"github.com/zeabur/zbpack/pkg/types"
)

// PackageManager defines an interface for common package management operations.
// It is basically the reimplementation of Corepack + Ni (https://github.com/antfu-collective/ni)
type PackageManager interface {
	GetType() types.NodePackageManager
	GetInitCommand() string
	GetInstallProjectDependenciesCommand() string
	GetRunScript(script string) string
}

// Npm is the implementation of PackageManager for npm.
type Npm struct {
	MajorVersion uint64
}

var _ PackageManager = Npm{}

// GetType returns the type of the package manager.
func (Npm) GetType() types.NodePackageManager {
	return types.NodePackageManagerNpm
}

// GetInitCommand returns the command to install npm.
func (n Npm) GetInitCommand() string {
	if n.MajorVersion == 0 {
		return "npm update -g npm"
	}

	return fmt.Sprintf("npm install -g npm@%d", n.MajorVersion)
}

// GetInstallProjectDependenciesCommand returns the command to install project dependencies.
func (Npm) GetInstallProjectDependenciesCommand() string {
	return "npm install"
}

// GetRunScript returns the command to run a script.
func (Npm) GetRunScript(script string) string {
	return "npm run " + script
}

// Yarn is the implementation of PackageManager for yarn.
type Yarn struct {
	MajorVersion uint64
}

var _ PackageManager = Yarn{}

// GetType returns the type of the package manager.
func (Yarn) GetType() types.NodePackageManager {
	return types.NodePackageManagerYarn
}

// GetInitCommand returns the command to install yarn.
func (y Yarn) GetInitCommand() string {
	command := "npm install -g yarn@latest"

	if y.MajorVersion > 1 { // berry
		command += " && yarn set version berry"
	}

	return command
}

// GetInstallProjectDependenciesCommand returns the command to install project dependencies.
func (Yarn) GetInstallProjectDependenciesCommand() string {
	return "yarn install"
}

// GetRunScript returns the command to run a script.
func (Yarn) GetRunScript(script string) string {
	return "yarn " + script
}

// Pnpm is the implementation of PackageManager for pnpm.
type Pnpm struct {
	MajorVersion uint64
}

var _ PackageManager = Pnpm{}

// GetType returns the type of the package manager.
func (Pnpm) GetType() types.NodePackageManager {
	return types.NodePackageManagerPnpm
}

// GetInitCommand returns the command to install pnpm.
func (p Pnpm) GetInitCommand() string {
	if p.MajorVersion == 0 {
		return "npm install -g pnpm@latest || npm install -g pnpm@8"
	}

	return fmt.Sprintf("npm install -g pnpm@%d", p.MajorVersion)
}

// GetInstallProjectDependenciesCommand returns the command to install project dependencies.
func (Pnpm) GetInstallProjectDependenciesCommand() string {
	return "pnpm install"
}

// GetRunScript returns the command to run a script.
func (Pnpm) GetRunScript(script string) string {
	return "pnpm " + script
}

// Bun is the implementation of PackageManager for bun.
type Bun struct{}

var _ PackageManager = Bun{}

// GetType returns the type of the package manager.
func (Bun) GetType() types.NodePackageManager {
	return types.NodePackageManagerBun
}

// GetInitCommand returns the command to install bun.
func (Bun) GetInitCommand() string {
	return "npm install -g bun@latest"
}

// GetInstallProjectDependenciesCommand returns the command to install project dependencies.
func (Bun) GetInstallProjectDependenciesCommand() string {
	return "bun install"
}

// GetRunScript returns the command to run a script.
func (Bun) GetRunScript(script string) string {
	return "bun run " + script
}

// UnspecifiedPackageManager is the implementation of PackageManager
// for an unspecified package manager.
//
// The type will be set to unknown.
type UnspecifiedPackageManager struct {
	PackageManager
}

// GetType returns the type of the package manager.
func (u UnspecifiedPackageManager) GetType() types.NodePackageManager {
	return types.NodePackageManagerUnknown
}
