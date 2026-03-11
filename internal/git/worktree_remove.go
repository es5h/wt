package git

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/es5h/wt/internal/runner"
)

func WorktreeRemove(ctx context.Context, r runner.Runner, repoRoot string, path string, force bool) error {
	resolvedPath, err := resolveRemovalTarget(repoRoot, path)
	if err != nil {
		return err
	}

	if err := makeWorktreeWritable(resolvedPath); err != nil {
		return fmt.Errorf("prepare worktree remove target %s: %w", resolvedPath, err)
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, resolvedPath)

	res, err := r.Run(ctx, repoRoot, "git", args...)
	if err != nil {
		msg := commandError(res, err)
		if strings.Contains(strings.ToLower(msg), "permission denied") {
			msg = fmt.Sprintf("%s (target=%s; %s)", msg, resolvedPath, diagnoseRemovalFailure(resolvedPath))
		}
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return nil
}

func resolveRemovalTarget(repoRoot string, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		repoRoot = strings.TrimSpace(repoRoot)
		if repoRoot == "" {
			return "", fmt.Errorf("repo root is required for relative remove path")
		}
		clean = filepath.Join(repoRoot, clean)
	}

	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("resolve remove path: %w", err)
	}
	if abs == string(filepath.Separator) {
		return "", fmt.Errorf("refusing to remove filesystem root")
	}
	return abs, nil
}

func makeWorktreeWritable(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("target is not a directory")
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		mode := d.Type()
		if info, err := d.Info(); err == nil {
			mode = info.Mode()
		} else {
			return fmt.Errorf("stat %s: %w", path, err)
		}

		if mode&0o200 != 0 {
			return nil
		}
		if err := os.Chmod(path, mode|0o200); err != nil {
			return fmt.Errorf("chmod %s: %w", path, err)
		}
		return nil
	})
}

func diagnoseRemovalFailure(root string) string {
	info, err := os.Lstat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "worktree path no longer exists"
		}
		return fmt.Sprintf("cannot stat worktree path: %v", err)
	}
	if !info.IsDir() {
		return "worktree path is not a directory"
	}

	diag := "worktree path still exists after git remove"
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			diag = fmt.Sprintf("cannot access %s: %v", path, walkErr)
			return fs.SkipAll
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			diag = fmt.Sprintf("cannot stat %s: %v", path, err)
			return fs.SkipAll
		}
		if info.Mode()&0o200 == 0 {
			diag = fmt.Sprintf("non-writable entry remains: %s (mode=%#o)", path, info.Mode().Perm())
			return fs.SkipAll
		}
		return nil
	})
	return diag
}
