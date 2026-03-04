package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/worktree"
)

func newPathCmd() *cobra.Command {
	var jsonOut bool
	var create bool
	var tui bool
	var noTui bool
	var createPath string
	var createRoot string
	var createFrom string

	cmd := &cobra.Command{
		Use:   "path <query>",
		Short: "Print selected worktree path",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completePathQuery(cmd, args, toComplete)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tui {
				return usageError(fmt.Errorf("wt path: --tui is not implemented yet"))
			}
			_ = noTui

			query := strings.TrimSpace(args[0])
			if query == "" {
				return usageError(fmt.Errorf("wt path: query cannot be empty"))
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

			wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			matches := matchWorktrees(wts, query)
			if len(matches) == 0 {
				if !create {
					return &exitError{Code: 1, Err: fmt.Errorf("wt path: no matches for %q", query)}
				}

				branch := query
				if strings.TrimSpace(createFrom) == "" {
					if after, ok := strings.CutPrefix(branch, "origin/"); ok && strings.TrimSpace(after) != "" {
						branch = after
					}
				}

				primaryRoot, err := git.PrimaryWorktreeRoot(ctx, d.Runner, repoRoot)
				if err != nil {
					return err
				}

				path, err := createWorktreeFromList(ctx, d, repoRoot, primaryRoot, branch, createOpts{
					Path: createPath,
					Root: createRoot,
					From: createFrom,
				}, wts)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), path)
				return nil
			}
			chosen, err := resolveMatchedWorktree("wt path", wts, query)
			if err != nil {
				return err
			}
			if jsonOut {
				type out struct {
					Path   string `json:"path"`
					Branch string `json:"branch,omitempty"`
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out{
					Path:   chosen.Path,
					Branch: strings.TrimPrefix(chosen.Branch, "refs/heads/"),
				})
			}

			// IMPORTANT: stdout must contain only the path (for: cd "$(wt path ...)")
			fmt.Fprintln(cmd.OutOrStdout(), chosen.Path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	cmd.Flags().BoolVar(&create, "create", false, "create worktree if missing")
	cmd.Flags().BoolVar(&tui, "tui", false, "use TUI selection (not implemented yet)")
	cmd.Flags().BoolVar(&noTui, "no-tui", false, "disable TUI selection (reserved)")
	cmd.Flags().StringVar(&createPath, "path", "", "worktree path for --create")
	cmd.Flags().StringVar(&createRoot, "root", "", "worktree root for --create default path resolution")
	cmd.Flags().StringVar(&createFrom, "from", "", "start point for --create (default: origin/<branch> if exists, else origin/HEAD or main)")

	return cmd
}

func completePathQuery(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	d, err := getDeps(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := cmd.Context()
	repoRoot, err := git.RepoRoot(ctx, d.Runner, d.Cwd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	prefix := strings.ToLower(strings.TrimSpace(toComplete))

	uniq := map[string]struct{}{}
	for _, wt := range wts {
		if wt.Branch != "" && !wt.Detached {
			branchShort := strings.TrimPrefix(wt.Branch, "refs/heads/")
			if branchShort != "" {
				uniq[branchShort] = struct{}{}
			}
			continue
		}

		base := filepath.Base(wt.Path)
		if base != "" {
			uniq[base] = struct{}{}
		}
	}

	if os.Getenv("WT_PATH_COMPLETE_REMOTE") == "1" {
		branches, err := git.RemoteBranches(ctx, d.Runner, repoRoot, "origin")
		if err == nil {
			for _, b := range branches {
				if b != "" {
					uniq[b] = struct{}{}
				}
			}
		}
	}

	out := make([]string, 0, len(uniq))
	for c := range uniq {
		if prefix == "" || strings.HasPrefix(strings.ToLower(c), prefix) {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out, cobra.ShellCompDirectiveNoFileComp
}

func matchWorktrees(wts []worktree.Worktree, query string) []worktree.Worktree {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	type scored struct {
		wt    worktree.Worktree
		score int
	}

	var scoredMatches []scored
	for _, wt := range wts {
		base := strings.ToLower(filepath.Base(wt.Path))
		path := strings.ToLower(wt.Path)

		branchShort := strings.TrimPrefix(wt.Branch, "refs/heads/")
		branchShortLower := strings.ToLower(branchShort)
		branchLower := strings.ToLower(wt.Branch)

		score := 0
		switch {
		case base == q:
			score = 300
		case branchShortLower == q && branchShort != "":
			score = 290
		case path == q:
			score = 280
		case branchLower == q && wt.Branch != "":
			score = 270
		case strings.Contains(base, q):
			score = 200
		case strings.Contains(branchShortLower, q) && branchShort != "":
			score = 190
		case strings.Contains(path, q):
			score = 150
		case strings.Contains(branchLower, q) && wt.Branch != "":
			score = 140
		default:
			score = 0
		}

		if score > 0 {
			scoredMatches = append(scoredMatches, scored{wt: wt, score: score})
		}
	}

	sort.SliceStable(scoredMatches, func(i, j int) bool {
		if scoredMatches[i].score != scoredMatches[j].score {
			return scoredMatches[i].score > scoredMatches[j].score
		}
		return scoredMatches[i].wt.Path < scoredMatches[j].wt.Path
	})

	matches := make([]worktree.Worktree, 0, len(scoredMatches))
	for _, m := range scoredMatches {
		matches = append(matches, m.wt)
	}
	return matches
}
