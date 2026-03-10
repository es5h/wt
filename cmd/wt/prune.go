package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/git"
	tuipicker "github.com/es5h/wt/internal/tui/picker"
	"github.com/es5h/wt/internal/worktree"
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
	var tui bool

	cmd := &cobra.Command{
		Use:           "prune",
		Short:         "Preview or remove stale prunable worktree entries",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tui && jsonOut {
				return usageError(fmt.Errorf("wt prune: --tui cannot be combined with --json"))
			}

			d, err := getDeps(cmd)
			if err != nil {
				return err
			}
			if tui && (d.CanUseTUI == nil || !d.CanUseTUI()) {
				return usageError(fmt.Errorf("wt prune: --tui requires a TTY on stdin and stderr"))
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
			if tui && len(candidates) > 0 {
				if err := previewPruneCandidatesWithTUI(cmd, d, candidates, apply); err != nil {
					return err
				}
			}
			if !apply {
				return writePruneOutput(cmd, candidates, false, jsonOut)
			}

			if len(candidates) == 0 {
				return writePruneOutput(cmd, candidates, true, jsonOut)
			}
			if tui {
				confirm := confirmPrune
				if d.ConfirmPrune != nil {
					confirm = d.ConfirmPrune
				}
				confirmed, err := confirm(cmd.InOrStdin(), cmd.ErrOrStderr(), len(candidates))
				if err != nil {
					return err
				}
				if !confirmed {
					return &exitError{Code: 1, Err: fmt.Errorf("wt prune: aborted")}
				}
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
	cmd.Flags().BoolVar(&tui, "tui", false, "preview prunable entries in TUI before returning or applying")
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

func previewPruneCandidatesWithTUI(cmd *cobra.Command, d *deps, items []pruneCandidate, apply bool) error {
	if d == nil || d.CanUseTUI == nil || !d.CanUseTUI() {
		return usageError(fmt.Errorf("wt prune: --tui requires a TTY on stdin and stderr"))
	}

	preview := previewPruneWithTUI
	if d.PreviewPrune != nil {
		preview = d.PreviewPrune
	}

	err := preview(cmd, items, apply)
	if err == nil {
		return nil
	}
	if errors.Is(err, tuipicker.ErrCancelled) {
		return &exitError{Code: 130, Err: fmt.Errorf("wt prune: preview cancelled")}
	}
	if errors.Is(err, tuipicker.ErrNonTTY) {
		return usageError(fmt.Errorf("wt prune: --tui requires a TTY on stdin and stderr"))
	}
	return err
}

func previewPruneWithTUI(cmd *cobra.Command, items []pruneCandidate, apply bool) error {
	in, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return tuipicker.ErrNonTTY
	}
	screen, ok := cmd.ErrOrStderr().(*os.File)
	if !ok {
		return tuipicker.ErrNonTTY
	}

	help := "Type to filter, arrows/Ctrl-N/Ctrl-P move, Enter close preview, Esc cancel"
	if apply {
		help = "Type to filter, arrows/Ctrl-N/Ctrl-P move, Enter continue to confirm prune, Esc cancel"
	}

	_, err := tuipicker.Run(in, screen, tuipicker.Config{
		Title: "Preview prune candidates",
		Help:  help,
		Items: buildPrunePickerItems(items),
	})
	return err
}

func buildPrunePickerItems(items []pruneCandidate) []tuipicker.Item {
	pickerItems := make([]tuipicker.Item, 0, len(items))
	for _, item := range items {
		label := item.Branch
		if label == "" {
			label = filepath.Base(item.Path)
		}

		metaParts := []string{"prunable"}
		if item.PruneReason != "" {
			metaParts = append(metaParts, item.PruneReason)
		}

		pickerItems = append(pickerItems, tuipicker.Item{
			ID:         item.Path,
			Label:      label,
			Detail:     item.Path,
			Meta:       strings.Join(metaParts, " | "),
			FilterText: strings.Join([]string{label, item.Path, item.Branch, item.PruneReason}, " "),
		})
	}
	return pickerItems
}

func confirmPrune(in io.Reader, out io.Writer, count int) (bool, error) {
	prompt := fmt.Sprintf("Prune %d stale worktree entr", count)
	if count == 1 {
		prompt += "y"
	} else {
		prompt += "ies"
	}
	prompt += " with git worktree prune --expire now? [y/N] "

	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("wt prune: failed to read confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
