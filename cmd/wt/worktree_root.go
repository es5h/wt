package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"wt/internal/git"
)

func resolveCreateTargetPath(ctx context.Context, d *deps, repoRoot string, branch string, opts createOpts) (string, error) {
	explicitPath := strings.TrimSpace(opts.Path)
	if explicitPath != "" {
		return explicitPath, nil
	}

	branchPathPart, err := safeBranchPathPart(branch)
	if err != nil {
		return "", usageError(err)
	}

	root, err := resolveWorktreeRoot(ctx, d, repoRoot, opts.Root)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, branchPathPart), nil
}

func resolveWorktreeRoot(ctx context.Context, d *deps, repoRoot string, cliRoot string) (string, error) {
	switch {
	case strings.TrimSpace(cliRoot) != "":
		return normalizeWorktreeRoot(repoRoot, cliRoot), nil
	case strings.TrimSpace(os.Getenv("WT_ROOT")) != "":
		return normalizeWorktreeRoot(repoRoot, os.Getenv("WT_ROOT")), nil
	}

	configRoot, err := git.ConfigGetLocal(ctx, d.Runner, repoRoot, "wt.root")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(configRoot) != "" {
		return normalizeWorktreeRoot(repoRoot, configRoot), nil
	}

	return filepath.Join(repoRoot, ".wt"), nil
}

func normalizeWorktreeRoot(repoRoot string, root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return filepath.Join(repoRoot, ".wt")
	}
	if filepath.IsAbs(root) {
		return filepath.Clean(root)
	}
	return filepath.Clean(filepath.Join(repoRoot, root))
}
