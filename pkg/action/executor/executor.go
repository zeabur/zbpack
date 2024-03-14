// Package executor installs modules and provides a function to run an action with some Zeabur Pack extensions.
package executor

import (
	"context"
	"fmt"

	zbaction "github.com/zeabur/action"
	"github.com/zeabur/action/environment"
	"github.com/zeabur/action/procedures/procvariables"

	// modules: action – mandatory
	_ "github.com/zeabur/action/procedures"
	_ "github.com/zeabur/action/procedures/artifact"

	// modules: zbpack extensions
	_ "github.com/zeabur/zbpack/internal/action"
	_ "github.com/zeabur/zbpack/internal/golang/proc"
	_ "github.com/zeabur/zbpack/internal/python/proc"

	// module: action – environment
	_ "github.com/zeabur/zbpack/internal/golang/env"
	_ "github.com/zeabur/zbpack/internal/python/env"
)

// ValidateEnvironment checks if the environment matches the requirements.
func ValidateEnvironment(action zbaction.Action) error {
	requirement, err := zbaction.CompileActionRequirement(action)
	if err != nil {
		return fmt.Errorf("compile action requirement: %w", err)
	}

	softwareList := environment.DetermineSoftwareList()

	if err := requirement.CheckRequirement(softwareList); err != nil {
		return fmt.Errorf("requirement not met: %w", err)
	}

	return nil
}

// RunAction runs the given action with arguments.
func RunAction(ctx context.Context, action zbaction.Action, options ...zbaction.ExecutorOptionsFn) error {
	allOptions := append([]zbaction.ExecutorOptionsFn{
		procvariables.WithEnvBuildkitHost(),
	}, options...)

	return zbaction.RunAction(
		ctx, action,
		allOptions...,
	)
}
