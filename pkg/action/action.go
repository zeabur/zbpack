// Package action provides the procedure arguments for Zeabur Pack procedures and the executor of Zeabur Pack.
package action

import (
	zbaction "github.com/zeabur/action"
	"github.com/zeabur/action/procedures/procvariables"
)

// Arguments is a procedure argument injected when triggering RunAction.
//
// Any data type except for string must be converted to string.
// For numeric types (int, float, etc.), it should be in the decimal format (1 or 1.0).
// For boolean types, it should be "true" or "false".
type Arguments map[ArgumentKey]string

// ArgumentKey is the argument key for Zeabur Pack procedures.
type ArgumentKey = string

const (
	// ArgContainerPush (boolean) indicates if we should push the container to the registry.
	ArgContainerPush ArgumentKey = "zbpack.container.push"

	// ArgContainerTag (string) is the tag of the container image.
	//
	// This is the tag of the image that will be pushed to.
	// For example, `docker.io/zeabur/service-12345678:latest`.
	ArgContainerTag ArgumentKey = "zbpack.container.image"
)

const (
	// ArgGitRepo (string) is the URL of the Git repository.
	ArgGitRepo ArgumentKey = "zbpack.git.repo"
	// ArgGitBranch (string) is the branch of the Git repository.
	ArgGitBranch ArgumentKey = "zbpack.git.branch"
	// ArgGitDepth (int) is the depth of the Git repository.
	ArgGitDepth ArgumentKey = "zbpack.git.depth"
	// ArgGitAuthUsername (string) is the username for the Git repository.
	ArgGitAuthUsername ArgumentKey = "zbpack.git.auth.username"
	// ArgGitAuthPassword (string) is the password for the Git repository.
	ArgGitAuthPassword ArgumentKey = "zbpack.git.auth.password"

	// ArgLocalPath (string) is the local path to the repository.
	ArgLocalPath ArgumentKey = "zbpack.local.path"
)

// WithArg is a function to inject an argument to the procedure.
func WithArg(key ArgumentKey, value string) zbaction.ExecutorOptionsFn {
	return procvariables.WithProcVariables(key, value)
}
