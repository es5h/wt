package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wt/internal/git"
)

func resolveCreateTargetPath(ctx context.Context, d *deps, repoRoot string, primaryRoot string, branch string, opts createOpts) (string, error) {
	explicitPath := strings.TrimSpace(opts.Path)
	if explicitPath != "" {
		return explicitPath, nil
	}

	branchPathPart, err := safeBranchPathPart(branch)
	if err != nil {
		return "", usageError(err)
	}

	root, err := resolveWorktreeRoot(ctx, d, repoRoot, primaryRoot, opts.Root)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, branchPathPart), nil
}

func resolveWorktreeRoot(ctx context.Context, d *deps, repoRoot string, primaryRoot string, cliRoot string) (string, error) {
	if strings.TrimSpace(primaryRoot) == "" {
		return "", fmt.Errorf("missing primary root (internal error)")
	}

	switch {
	case strings.TrimSpace(cliRoot) != "":
		return normalizeWorktreeRoot(primaryRoot, cliRoot), nil
	case strings.TrimSpace(os.Getenv("WT_ROOT")) != "":
		return normalizeWorktreeRoot(primaryRoot, os.Getenv("WT_ROOT")), nil
	}

	configRoot, err := git.ConfigGetLocal(ctx, d.Runner, repoRoot, "wt.root")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(configRoot) != "" {
		return normalizeWorktreeRoot(primaryRoot, configRoot), nil
	}

	return filepath.Join(primaryRoot, ".wt"), nil
}

func normalizeWorktreeRoot(baseRoot string, root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return filepath.Join(baseRoot, ".wt")
	}
	if filepath.IsAbs(root) {
		return filepath.Clean(root)
	}
	return filepath.Clean(filepath.Join(baseRoot, root))
}
