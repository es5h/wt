package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/es5h/wt/internal/runner"
)

func CommonDir(ctx context.Context, r runner.Runner, workDir string) (string, error) {
	res, err := r.Run(ctx, workDir, "git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --git-common-dir: %s", commandError(res, err))
	}
	out := strings.TrimSpace(string(res.Stdout))
	if out == "" {
		return "", fmt.Errorf("git rev-parse --git-common-dir: empty output")
	}
	return out, nil
}

// PrimaryWorktreeRoot returns the root directory of the "primary" worktree (the one
// that contains the common git directory). This is stable across linked worktrees,
// and prevents accidental nested worktree creation (e.g. <wt>/.wt/<a>/.wt/<b>...).
func PrimaryWorktreeRoot(ctx context.Context, r runner.Runner, workDir string) (string, error) {
	commonDir, err := CommonDir(ctx, r, workDir)
	if err != nil {
		return "", err
	}

	// In normal non-bare repos, commonDir ends with "/.git".
	// For worktrees, it points back to the main repo's ".git".
	root := filepath.Dir(commonDir)
	if root == "." || strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("git rev-parse --git-common-dir: unexpected output: %q", commonDir)
	}
	return root, nil
}
