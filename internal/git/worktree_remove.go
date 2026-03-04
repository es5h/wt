package git

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"wt/internal/runner"
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
	if err == nil {
		return nil
	}

	if force && isPermissionDenied(res, err) {
		if chmodErr := makeUserWritable(path); chmodErr != nil {
			return fmt.Errorf("git %s: %s (failed to make worktree writable: %v)", strings.Join(args, " "), commandError(res, err), chmodErr)
		}

		res, err = r.Run(ctx, repoRoot, "git", args...)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("git %s: %s", strings.Join(args, " "), commandError(res, err))
}

func isPermissionDenied(res runner.Result, err error) bool {
	msg := strings.ToLower(strings.TrimSpace(string(res.Stderr) + "\n" + err.Error()))
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "허가 거부")
}

func makeUserWritable(path string) error {
	return filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		mode := info.Mode()
		if mode.Perm()&0o200 != 0 {
			return nil
		}

		return os.Chmod(current, mode.Perm()|0o200)
	})
}
