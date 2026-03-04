package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/worktree"
)

func newRunCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "run <query> -- <cmd...>",
		Short: "Run a command in a selected worktree",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return usageError(fmt.Errorf("wt run: requires <query> and <cmd...> (use: wt run <query> -- <cmd...>)"))
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeGotoQuery(cmd, args, toComplete)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return usageError(fmt.Errorf("wt run: query cannot be empty"))
			}

			command := append([]string(nil), args[1:]...)
			if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
				return usageError(fmt.Errorf("wt run: command cannot be empty"))
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

			chosen, err := resolveMatchedWorktree("wt run", wts, query)
			if err != nil {
				return err
			}

			res, err := d.Runner.Run(ctx, chosen.Path, command[0], command[1:]...)
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if encodeErr := enc.Encode(struct {
					Path     string   `json:"path"`
					Command  []string `json:"command"`
					ExitCode int      `json:"exitCode"`
				}{
					Path:     chosen.Path,
					Command:  command,
					ExitCode: res.ExitCode,
				}); encodeErr != nil {
					return encodeErr
				}
			} else {
				if len(res.Stdout) > 0 {
					if _, writeErr := cmd.OutOrStdout().Write(res.Stdout); writeErr != nil {
						return writeErr
					}
				}
				if len(res.Stderr) > 0 {
					if _, writeErr := cmd.ErrOrStderr().Write(res.Stderr); writeErr != nil {
						return writeErr
					}
				}
			}

			if err != nil {
				if res.ExitCode != 0 {
					return &exitError{Code: res.ExitCode}
				}
				return fmt.Errorf("wt run: %w", err)
			}
			if res.ExitCode != 0 {
				return &exitError{Code: res.ExitCode}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}

func resolveMatchedWorktree(commandName string, wts []worktree.Worktree, query string) (worktree.Worktree, error) {
	matches := matchWorktrees(wts, query)
	if len(matches) == 0 {
		return worktree.Worktree{}, &exitError{Code: 1, Err: fmt.Errorf("%s: no matches for %q", commandName, query)}
	}
	if len(matches) > 1 {
		return worktree.Worktree{}, &exitError{Code: 1, Err: fmt.Errorf("%s", formatAmbiguousSelection(commandName, query, matches))}
	}
	return matches[0], nil
}

func formatAmbiguousSelection(commandName string, query string, matches []worktree.Worktree) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %d matches for %q\n", commandName, len(matches), query)
	for _, wt := range matches {
		fmt.Fprintf(&b, "  - %s (%s)\n", wt.Path, displayBranch(wt))
	}
	b.WriteString("hint: use a more specific query (TUI selection is not implemented yet)")
	return b.String()
}
