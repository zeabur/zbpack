package zeaburpack

import (
	"fmt"
	"github.com/zeabur/zbpack/internal/source"
	"strings"

	"github.com/zeabur/zbpack/pkg/types"
)

const (
	reset  = "\033[0m"
	yellow = "\033[0;33m"
	blue   = "\033[0;34m"
)

// PrintPlanAndMeta prints the build plan and meta in a table format.
func PrintPlanAndMeta(plan types.PlanType, meta types.PlanMeta, printFunc func(log string)) {
	table := fmt.Sprintf(
		"\n%s╔══════════════════════════ %s%s %s═════════════════════════╗\n",
		blue, yellow, "Build Plan", blue,
	)

	table += fmt.Sprintf(
		"%s║%s %-16s %s│%s %-42s %s║%s\n",
		blue, reset, "provider", blue, reset, string(plan), blue, reset,
	)

	for k, v := range meta {
		if v == "" || v == "false" {
			continue
		}
		table += blue + "║───────────────────────────────────────────────────────────────║\n" + reset
		table += fmt.Sprintf(
			"%s║%s %-16s %s│%s %-42s %s║\n%s",
			blue, reset, k, blue, reset, v, blue, reset,
		)
	}

	table += fmt.Sprintf(
		"%s╚═══════════════════════════════════════════════════════════════╝%s\n",
		blue, reset,
	)

	printFunc(table)
}

// getGitHubSourceFromUrl returns a GitHub source from a GitHub URL.
func getGitHubSourceFromUrl(url, token string) (*source.Source, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		panic("Invalid GitHub URL")
	}
	repoOwner := parts[3]
	repoName := parts[4]
	src := source.NewGitHubSource(repoOwner, repoName, token)
	return &src, nil
}
