package transformer

import (
	"fmt"
	"os/exec"

	"github.com/zeabur/zbpack/pkg/types"
)

// TransformNix push the Nix Docker image output to registry.
func TransformNix(ctx *Context) error {
	if ctx.PlanType != types.PlanTypeNix {
		return ErrSkip
	}

	ctx.Log("Transforming Nix...")

	dockerTar := ctx.BuildkitPath

	if !ctx.PushImage {
		// SAFE: zbpack are managed by ourselves. Besides,
		// macOS does not contain policy.json by default.
		skopeoCmd := exec.Command("skopeo", "copy", "--insecure-policy", "docker-archive:"+dockerTar.Name(), "docker-daemon:"+ctx.ResultImage+":latest")
		skopeoCmd.Stdout = ctx.LogWriter
		skopeoCmd.Stderr = ctx.LogWriter
		if err := skopeoCmd.Run(); err != nil {
			return fmt.Errorf("run skopeo copy: %w", err)
		}
	} else {
		// SAFE: zbpack are managed by ourselves. Besides,
		// macOS does not contain policy.json by default.
		skopeoCmd := exec.Command("skopeo", "copy", "--insecure-policy", "docker-archive:"+dockerTar.Name(), "docker://"+ctx.ResultImage)
		skopeoCmd.Stdout = ctx.LogWriter
		skopeoCmd.Stderr = ctx.LogWriter
		if err := skopeoCmd.Run(); err != nil {
			return fmt.Errorf("run skopeo copy: %w", err)
		}
	}

	// remove the TAR since we have imported it
	return dockerTar.Remove("")
}
