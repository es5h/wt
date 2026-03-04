package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/worktree"
)

type pruneCandidate struct {
	Path        string `json:"path"`
	Branch      string `json:"branch,omitempty"`
	PruneReason string `json:"pruneReason,omitempty"`
	Action      string `json:"action"`
	Removed     bool   `json:"removed"`
}

func newPruneCmd() *cobra.Command {
	var jsonOut bool
	var apply bool

	cmd := &cobra.Command{
		Use:           "prune",
		Short:         "Preview or remove stale prunable worktree entries",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := getDeps(cmd)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			repoRoot, err := git.RepoRoot(ctx, d.Runner, d.Cwd)
			if err != nil {
				return err
			}

			before, err := git.WorktreeList(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			candidates := collectPrunableCandidates(before)
			if !apply {
				return writePruneOutput(cmd, candidates, false, jsonOut)
			}

			if len(candidates) == 0 {
				return writePruneOutput(cmd, candidates, true, jsonOut)
			}

			if err := git.WorktreePrune(ctx, d.Runner, repoRoot); err != nil {
				return err
			}

			after, err := git.WorktreeList(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			removed := removedPruneCandidates(candidates, after)
			return writePruneOutput(cmd, removed, true, jsonOut)
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "actually run git worktree prune --expire now")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}

func collectPrunableCandidates(wts []worktree.Worktree) []pruneCandidate {
	out := make([]pruneCandidate, 0)
	for _, wt := range wts {
		if !wt.Prunable {
			continue
		}
		out = append(out, pruneCandidate{
			Path:        wt.Path,
			Branch:      strings.TrimPrefix(wt.Branch, "refs/heads/"),
			PruneReason: wt.PruneReason,
			Action:      "preview",
			Removed:     false,
		})
	}
	return out
}

func removedPruneCandidates(before []pruneCandidate, after []worktree.Worktree) []pruneCandidate {
	remaining := make(map[string]struct{}, len(after))
	for _, wt := range after {
		remaining[wt.Path] = struct{}{}
	}

	out := make([]pruneCandidate, 0, len(before))
	for _, item := range before {
		item.Action = "prune"
		_, ok := remaining[item.Path]
		item.Removed = !ok
		out = append(out, item)
	}
	return out
}

func writePruneOutput(cmd *cobra.Command, items []pruneCandidate, applied bool, jsonOut bool) error {
	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}

	for _, item := range items {
		action := "would-prune"
		if applied {
			if item.Removed {
				action = "pruned"
			} else {
				action = "kept"
			}
		}

		line := fmt.Sprintf("%s  %s", action, item.Path)
		if item.Branch != "" {
			line += fmt.Sprintf("  (%s)", item.Branch)
		}
		if item.PruneReason != "" {
			line += fmt.Sprintf("  [%s]", item.PruneReason)
		}
		fmt.Fprintln(cmd.OutOrStdout(), line)
	}
	return nil
}
