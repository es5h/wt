package git

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"wt/internal/runner"
	"wt/internal/worktree"
)

func RepoRoot(ctx context.Context, r runner.Runner, cwd string) (string, error) {
	res, err := r.Run(ctx, cwd, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %s", commandError(res, err))
	}
	root := strings.TrimSpace(string(res.Stdout))
	if root == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel: empty output")
	}
	return root, nil
}

func WorktreeListPorcelain(ctx context.Context, r runner.Runner, repoRoot string) ([]byte, error) {
	res, err := r.Run(ctx, repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git worktree list --porcelain: %s", commandError(res, err))
	}
	return res.Stdout, nil
}

func WorktreeList(ctx context.Context, r runner.Runner, repoRoot string) ([]worktree.Worktree, error) {
	out, err := WorktreeListPorcelain(ctx, r, repoRoot)
	if err != nil {
		return nil, err
	}
	wts, err := worktree.ParsePorcelain(bytes.NewReader(out))
	if err != nil {
		return nil, fmt.Errorf("parse worktree porcelain: %w", err)
	}
	return wts, nil
}

func DefaultBaseRef(ctx context.Context, r runner.Runner, repoRoot string) string {
	res, err := r.Run(ctx, repoRoot, "git", "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(string(res.Stdout))
		const prefix = "refs/remotes/"
		if after, ok := strings.CutPrefix(ref, prefix); ok {
			ref = after
		}
		if ref != "" {
			return ref
		}
	}
	return "main"
}

func RefExists(ctx context.Context, r runner.Runner, repoRoot string, ref string) (bool, error) {
	if strings.TrimSpace(ref) == "" {
		return false, fmt.Errorf("empty ref")
	}

	res, err := r.Run(ctx, repoRoot, "git", "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err == nil {
		return true, nil
	}
	if res.ExitCode == 1 {
		return false, nil
	}
	return false, fmt.Errorf("git rev-parse --verify %s: %s", ref, commandError(res, err))
}

func ConfigGetLocal(ctx context.Context, r runner.Runner, repoRoot string, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("empty config key")
	}

	res, err := r.Run(ctx, repoRoot, "git", "config", "--local", "--get", key)
	if err == nil {
		return strings.TrimSpace(string(res.Stdout)), nil
	}
	if res.ExitCode == 1 {
		return "", nil
	}
	return "", fmt.Errorf("git config --local --get %s: %s", key, commandError(res, err))
}

func IsAncestor(ctx context.Context, r runner.Runner, repoRoot string, olderRef string, newerRef string) (bool, error) {
	res, err := r.Run(ctx, repoRoot, "git", "merge-base", "--is-ancestor", olderRef, newerRef)
	if err == nil {
		return true, nil
	}
	if res.ExitCode == 1 {
		return false, nil
	}
	return false, fmt.Errorf("git merge-base --is-ancestor %s %s: %s", olderRef, newerRef, commandError(res, err))
}

func commandError(res runner.Result, err error) string {
	msg := strings.TrimSpace(string(res.Stderr))
	if msg != "" {
		return msg
	}
	if err != nil {
		return err.Error()
	}
	return "unknown error"
}
