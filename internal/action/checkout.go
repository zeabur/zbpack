package action

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/zbpack/pkg/action"
)

var (
	// ErrNoRepoProvided is an error indicating that no repository is provided.
	ErrNoRepoProvided = errors.New("no repository provided")
	// ErrNoBranchProvided is an error indicating that no branch is provided.
	ErrNoBranchProvided = errors.New("repository provided but no branch provided")
	// ErrNoLocalProvided is an error indicating that no local source is provided.
	ErrNoLocalProvided = errors.New("no local source provided")
	// ErrNoSourceProvided is an error indicating that no source is provided.
	ErrNoSourceProvided = errors.New("no any source provided")
)

func init() {
	zbaction.RegisterProcedure("zbpack/checkout", func(_ zbaction.ProcStepArgs) (zbaction.ProcedureStep, error) {
		return CheckoutProcedure{}, nil
	})
}

// CheckoutProcedure is a procedure that fetches the source code from the remote repository
// or the local file system according to the environment.
type CheckoutProcedure struct{}

// Run checkouts the codebase.
func (g CheckoutProcedure) Run(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	cleanup, err := g.checkoutFromLocal(ctx, sc)
	if err != nil {
		if !errors.Is(err, ErrNoLocalProvided) {
			return cleanup, fmt.Errorf("checkout from local: %w", err)
		}
	} else {
		return cleanup, nil
	}

	cleanup, err = g.checkoutFromGit(ctx, sc)
	if err != nil {
		if !errors.Is(err, ErrNoRepoProvided) {
			return cleanup, fmt.Errorf("checkout from git: %w", err)
		}
	} else {
		return cleanup, nil
	}

	return nil, ErrNoSourceProvided
}

func (g CheckoutProcedure) checkoutFromLocal(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	vars := sc.VariableContainer()

	localArgs := zbaction.ProcStepArgs{
		"dest": "/",
	}

	if local, ok := vars.GetVariable(action.ArgLocalPath); ok {
		localArgs["src"] = local
	} else {
		return nil, ErrNoLocalProvided
	}

	// construct an `action/copy-local-dir` procedure
	localProc, err := zbaction.ResolveProcedure("action/copy-local-dir", localArgs)
	if err != nil {
		return nil, fmt.Errorf("resolve action/copy-local-dir: %w", err)
	}

	// run the `action/copy-local-dir` procedure
	return localProc.Run(ctx, sc)
}

func (g CheckoutProcedure) checkoutFromGit(ctx context.Context, sc *zbaction.StepContext) (zbaction.CleanupFn, error) {
	vars := sc.VariableContainer()

	checkoutArgs := zbaction.ProcStepArgs{}

	if repo, ok := vars.GetVariable(action.ArgGitRepo); ok {
		checkoutArgs["url"] = repo
	} else {
		return nil, ErrNoRepoProvided
	}

	if branch, ok := vars.GetVariable(action.ArgGitBranch); ok {
		checkoutArgs["branch"] = branch
	} else {
		return nil, ErrNoBranchProvided
	}

	if depth, ok := vars.GetVariable(action.ArgGitDepth); ok {
		checkoutArgs["depth"] = depth
	}

	if authUsername, ok := vars.GetVariable(action.ArgGitAuthUsername); ok {
		checkoutArgs["authUsername"] = authUsername
	}
	if authPassword, ok := vars.GetVariable(action.ArgGitAuthPassword); ok {
		checkoutArgs["authPassword"] = authPassword
	}

	// construct an `action/checkout` procedure
	checkoutProc, err := zbaction.ResolveProcedure("action/checkout", checkoutArgs)
	if err != nil {
		return nil, fmt.Errorf("resolve action/checkout: %w", err)
	}

	// run the `action/checkout` procedure
	return checkoutProc.Run(ctx, sc)
}
