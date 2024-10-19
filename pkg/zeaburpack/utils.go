package zeaburpack

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/source"

	"github.com/zeabur/zbpack/pkg/types"
)

const (
	reset  = "\033[0m"
	yellow = "\033[0;33m"
	blue   = "\033[0;34m"
)

// PrintPlanAndMeta prints the build plan and meta in a table format.
func PrintPlanAndMeta(plan types.PlanType, meta types.PlanMeta, writer io.Writer) {
	table := fmt.Sprintf(
		"\n%s╔══════════════════════════════ %s%s %s═════════════════════════════╗\n",
		blue, yellow, "Build Plan", blue,
	)

	table += fmt.Sprintf(
		"%s║%s %-16s %s│%s %-50s %s║%s\n",
		blue, reset, "provider", blue, reset, string(plan), blue, reset,
	)

	for k, v := range meta {
		if v == "" || v == "false" {
			continue
		}
		table += blue + "║───────────────────────────────────────────────────────────────────────║\n" + reset
		if strings.Contains(v, "\n") {
			lines := strings.Split(v, "\n")
			for i, line := range lines {
				if i == 0 {
					table += fmt.Sprintf(
						"%s║%s %-16s %s│%s %-50s %s║\n",
						blue, reset, k, blue, reset, line, blue,
					)
					continue
				}
				table += fmt.Sprintf(
					"%s║%s %-16s %s│%s %-50s %s║\n",
					blue, reset, "", blue, reset, line, blue,
				)
			}
		} else {
			table += fmt.Sprintf(
				"%s║%s %-16s %s│%s %-50s %s║\n%s",
				blue, reset, k, blue, reset, v, blue, reset,
			)
		}
	}

	table += fmt.Sprintf(
		"%s╚═══════════════════════════════════════════════════════════════════════╝%s\n",
		blue, reset,
	)

	_, _ = writer.Write([]byte(table))
}

// getGitHubSourceFromURL returns a GitHub source from a GitHub URL.
func getGitHubSourceFromURL(url string, token *string) (afero.Fs, error) {
	repoAddress, ref, _ := strings.Cut(url, "#")
	parts := strings.Split(repoAddress, "/")
	if len(parts) < 5 {
		return nil, errors.New("invalid GitHub URL")
	}
	repoOwner := parts[3]
	repoName := parts[4]

	if ref != "" {
		return source.NewGitHubFs(repoOwner, repoName, token, source.GitHubRef(ref))
	}

	return source.NewGitHubFs(repoOwner, repoName, token)
}

func getS3SourceFromURL(url string, cfg *aws.Config) afero.Fs {
	return source.NewS3Fs(url, cfg)
}
