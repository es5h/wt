package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/es5h/wt/internal/runner"
)

func WorktreeRemove(ctx context.Context, r runner.Runner, repoRoot string, path string, force bool) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty path")
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	res, err := r.Run(ctx, repoRoot, "git", args...)
	if err != nil {
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), commandError(res, err))
	}
	return nil
}
