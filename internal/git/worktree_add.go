package git

import (
	"context"
	"fmt"
	"strings"

	"wt/internal/runner"
)

func WorktreeAddExistingBranch(ctx context.Context, r runner.Runner, repoRoot string, path string, branch string) error {
	path = strings.TrimSpace(path)
	branch = strings.TrimSpace(branch)
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if branch == "" {
		return fmt.Errorf("empty branch")
	}

	res, err := r.Run(ctx, repoRoot, "git", "worktree", "add", path, branch)
	if err != nil {
		return fmt.Errorf("git worktree add %s %s: %s", path, branch, commandError(res, err))
	}
	return nil
}

func WorktreeAddNewBranch(ctx context.Context, r runner.Runner, repoRoot string, path string, branch string, startPoint string) error {
	path = strings.TrimSpace(path)
	branch = strings.TrimSpace(branch)
	startPoint = strings.TrimSpace(startPoint)
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if branch == "" {
		return fmt.Errorf("empty branch")
	}
	if startPoint == "" {
		return fmt.Errorf("empty start point")
	}

	res, err := r.Run(ctx, repoRoot, "git", "worktree", "add", "-b", branch, path, startPoint)
	if err != nil {
		return fmt.Errorf("git worktree add -b %s %s %s: %s", branch, path, startPoint, commandError(res, err))
	}
	return nil
}
