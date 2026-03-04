package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
)

type removeResult struct {
	Path    string `json:"path"`
	Branch  string `json:"branch,omitempty"`
	Action  string `json:"action"`
	Removed bool   `json:"removed"`
}

func newRemoveCmd() *cobra.Command {
	var dryRun bool
	var force bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "remove <query>",
		Short: "Remove a selected worktree",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completePathQuery(cmd, args, toComplete)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return usageError(fmt.Errorf("wt remove: query cannot be empty"))
			}
			if !dryRun && !force {
				return usageError(fmt.Errorf("wt remove: requires --dry-run or --force"))
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
			primaryRoot, err := git.PrimaryWorktreeRoot(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			chosen, err := resolveMatchedWorktree("wt remove", wts, query)
			if err != nil {
				return err
			}
			if chosen.Path == repoRoot {
				return usageError(fmt.Errorf("wt remove: cannot remove current worktree: %s", chosen.Path))
			}
			if chosen.Path == primaryRoot {
				return usageError(fmt.Errorf("wt remove: cannot remove primary worktree: %s", chosen.Path))
			}
			if chosen.Prunable {
				return usageError(fmt.Errorf("wt remove: target is prunable; use 'wt prune --apply': %s", chosen.Path))
			}

			result := removeResult{
				Path:    chosen.Path,
				Branch:  strings.TrimPrefix(chosen.Branch, "refs/heads/"),
				Action:  "preview",
				Removed: false,
			}
			if !dryRun {
				if err := git.WorktreeRemove(ctx, d.Runner, repoRoot, chosen.Path, true); err != nil {
					return err
				}
				result.Action = "remove"
				result.Removed = true
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			status := "would-remove"
			if result.Removed {
				status = "removed"
			}
			line := fmt.Sprintf("%s  %s", status, result.Path)
			if result.Branch != "" {
				line += fmt.Sprintf("  (%s)", result.Branch)
			}
			fmt.Fprintln(cmd.OutOrStdout(), line)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview removal without changing anything")
	cmd.Flags().BoolVar(&force, "force", false, "actually remove the selected worktree")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}
