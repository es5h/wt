package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/worktree"
)

type createOpts struct {
	Path   string
	Root   string
	From   string
	DryRun bool
}

func newCreateCmd() *cobra.Command {
	var opts createOpts

	cmd := &cobra.Command{
		Use:           "create <branch>",
		Short:         "Create a new worktree for a branch",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := strings.TrimSpace(args[0])
			if branch == "" {
				return usageError(fmt.Errorf("wt create: branch cannot be empty"))
			}

			d, err := getDeps(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			repoRoot, err := git.RepoRoot(ctx, d.Runner, d.Cwd)
			if err != nil {
				return err
			}

			path, err := createWorktree(ctx, d, repoRoot, branch, opts)
			if err != nil {
				return err
			}

			// stdout must be path-only to allow scripting/cd chaining.
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Path, "path", "", "worktree path")
	cmd.Flags().StringVar(&opts.Root, "root", "", "worktree root for default path resolution")
	cmd.Flags().StringVar(&opts.From, "from", "", "start point for new branch (default: origin/HEAD or main)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print what would be executed (no changes)")
	return cmd
}

func createWorktree(ctx context.Context, d *deps, repoRoot string, branch string, opts createOpts) (string, error) {
	wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
	if err != nil {
		return "", err
	}
	return createWorktreeFromList(ctx, d, repoRoot, branch, opts, wts)
}

func createWorktreeFromList(ctx context.Context, d *deps, repoRoot string, branch string, opts createOpts, wts []worktree.Worktree) (string, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "", usageError(fmt.Errorf("wt create: branch cannot be empty"))
	}

	targetPath, err := resolveCreateTargetPath(ctx, d, repoRoot, branch, opts)
	if err != nil {
		return "", err
	}
	for _, wt := range wts {
		if wt.Branch == "refs/heads/"+branch && wt.Path != "" {
			return wt.Path, nil
		}
	}

	// If the local branch exists, we just add a worktree for it.
	localExists, err := git.RefExists(ctx, d.Runner, repoRoot, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}

	if localExists {
		if opts.DryRun {
			fmt.Fprintf(os.Stderr, "dry-run: git worktree add %s %s\n", targetPath, branch)
			return targetPath, nil
		}
		if err := git.WorktreeAddExistingBranch(ctx, d.Runner, repoRoot, targetPath, branch); err != nil {
			return "", err
		}
		return targetPath, nil
	}

	from := strings.TrimSpace(opts.From)
	if from == "" {
		from = git.DefaultBaseRef(ctx, d.Runner, repoRoot)
	}

	// If user explicitly provided --from, it takes precedence.
	startPoint := from
	if strings.TrimSpace(opts.From) == "" {
		remoteExists, err := git.RefExists(ctx, d.Runner, repoRoot, "refs/remotes/origin/"+branch)
		if err != nil {
			return "", err
		}
		if remoteExists {
			startPoint = "origin/" + branch
		}
	}

	exists, err := git.RefExists(ctx, d.Runner, repoRoot, startPoint)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", usageError(fmt.Errorf("wt create: start point does not exist: %s", startPoint))
	}

	if opts.DryRun {
		fmt.Fprintf(os.Stderr, "dry-run: git worktree add -b %s %s %s\n", branch, targetPath, startPoint)
		return targetPath, nil
	}

	if err := git.WorktreeAddNewBranch(ctx, d.Runner, repoRoot, targetPath, branch, startPoint); err != nil {
		return "", err
	}
	return targetPath, nil
}

func safeBranchPathPart(branch string) (string, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "", fmt.Errorf("wt create: branch cannot be empty")
	}

	p := filepath.Clean(filepath.FromSlash(branch))
	if p == "." || p == ".." || strings.HasPrefix(p, ".."+string(os.PathSeparator)) || filepath.IsAbs(p) {
		return "", fmt.Errorf("wt create: unsupported branch for default path: %q", branch)
	}
	return p, nil
}
