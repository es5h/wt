package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	tuipicker "github.com/es5h/wt/internal/tui/picker"
	"github.com/es5h/wt/internal/worktree"
)

func selectWorktreeWithTUI(cmd *cobra.Command, d *deps, commandName string, wts []worktree.Worktree, initialFilter string) (worktree.Worktree, error) {
	if d == nil || d.CanUseTUI == nil || !d.CanUseTUI() {
		return worktree.Worktree{}, usageError(fmt.Errorf("%s: --tui requires a TTY on stdin and stderr", commandName))
	}
	if len(wts) == 0 {
		return worktree.Worktree{}, &exitError{Code: 1, Err: fmt.Errorf("%s: no worktrees found", commandName)}
	}

	pick := pickWorktreeWithTUI
	if d.PickWorktree != nil {
		pick = d.PickWorktree
	}

	chosen, err := pick(cmd, wts, initialFilter)
	if err == nil {
		return chosen, nil
	}
	if errors.Is(err, tuipicker.ErrCancelled) {
		return worktree.Worktree{}, &exitError{Code: 130, Err: fmt.Errorf("%s: selection cancelled", commandName)}
	}
	if errors.Is(err, tuipicker.ErrNonTTY) {
		return worktree.Worktree{}, usageError(fmt.Errorf("%s: --tui requires a TTY on stdin and stderr", commandName))
	}
	return worktree.Worktree{}, err
}

func pickWorktreeWithTUI(cmd *cobra.Command, wts []worktree.Worktree, initialFilter string) (worktree.Worktree, error) {
	in, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return worktree.Worktree{}, tuipicker.ErrNonTTY
	}
	screen, ok := cmd.ErrOrStderr().(*os.File)
	if !ok {
		return worktree.Worktree{}, tuipicker.ErrNonTTY
	}

	items := make([]tuipicker.Item, 0, len(wts))
	byPath := make(map[string]worktree.Worktree, len(wts))
	for _, wt := range wts {
		metaParts := make([]string, 0, 3)
		if wt.HEAD != "" {
			metaParts = append(metaParts, shortCommit(wt.HEAD))
		}
		if wt.Locked {
			metaParts = append(metaParts, "locked")
		}
		if wt.Prunable {
			metaParts = append(metaParts, "prunable")
		}

		label := displayBranch(wt)
		if label == "(unknown)" {
			label = filepath.Base(wt.Path)
		}

		items = append(items, tuipicker.Item{
			ID:         wt.Path,
			Label:      label,
			Detail:     wt.Path,
			Meta:       strings.Join(metaParts, " "),
			FilterText: strings.Join([]string{label, wt.Path, wt.Branch, filepath.Base(wt.Path), wt.HEAD}, " "),
		})
		byPath[wt.Path] = wt
	}

	selected, err := tuipicker.Run(in, screen, tuipicker.Config{
		Title:         "Select worktree",
		Help:          "Type to filter, arrows/Ctrl-N/Ctrl-P move, Enter select, Esc cancel",
		Items:         items,
		InitialFilter: initialFilter,
	})
	if err != nil {
		return worktree.Worktree{}, err
	}

	chosen, ok := byPath[selected.ID]
	if !ok {
		return worktree.Worktree{}, fmt.Errorf("wt path: picker returned unknown worktree: %s", selected.ID)
	}
	return chosen, nil
}

func shortCommit(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}
