package git

import (
	"context"
	"fmt"
	"strings"

	"wt/internal/runner"
)

func RemoteBranches(ctx context.Context, r runner.Runner, repoRoot string, remote string) ([]string, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return nil, fmt.Errorf("empty remote")
	}

	res, err := r.Run(ctx, repoRoot, "git", "for-each-ref", "--format=%(refname:strip=3)", "refs/remotes/"+remote)
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref refs/remotes/%s: %s", remote, commandError(res, err))
	}

	lines := strings.Split(string(res.Stdout), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "HEAD" {
			continue
		}
		out = append(out, line)
	}
	return out, nil
}
