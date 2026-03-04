package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/hosting"
	"wt/internal/worktree"
)

type cleanupItem struct {
	Path              string `json:"path"`
	Branch            string `json:"branch,omitempty"`
	RecommendedAction string `json:"recommendedAction"`
	Action            string `json:"action"`
	Applied           bool   `json:"applied"`
	Removed           bool   `json:"removed"`
	Reason            string `json:"reason"`
	SafeToRemove      bool   `json:"safeToRemove"`

	MergedIntoBase   *bool  `json:"mergedIntoBase,omitempty"`
	BaseRef          string `json:"baseRef,omitempty"`
	MergedViaHosting *bool  `json:"mergedViaHosting,omitempty"`
	HostingProvider  string `json:"hostingProvider,omitempty"`
	HostingKind      string `json:"hostingKind,omitempty"`
}

type cleanupCandidate struct {
	Worktree worktree.Worktree
	Info     *verifyInfo
	Signals  listSignals
	Reason   string
}

func newCleanupCmd() *cobra.Command {
	var apply bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:           "cleanup",
		Short:         "Preview or apply recommended prune/remove actions",
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

			candidates, hostingNote, err := collectCleanupCandidates(ctx, d, repoRoot)
			if err != nil {
				return err
			}
			if hostingNote != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), hostingNote)
			}

			items, err := runCleanup(cmd, d, repoRoot, candidates, apply)
			if err != nil {
				return err
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(items)
			}

			for _, item := range items {
				fmt.Fprintln(cmd.OutOrStdout(), formatCleanupLine(item))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "actually run recommended prune/remove actions")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}

func collectCleanupCandidates(ctx context.Context, d *deps, repoRoot string) ([]cleanupCandidate, string, error) {
	wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
	if err != nil {
		return nil, "", err
	}

	baseRef := git.DefaultBaseRef(ctx, d.Runner, repoRoot)
	exists, err := git.RefExists(ctx, d.Runner, repoRoot, baseRef)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return nil, "", fmt.Errorf("wt cleanup: base ref does not exist: %s", baseRef)
	}

	verifyCtx := &listVerifyContext{
		RepoRoot:    repoRoot,
		BaseRef:     baseRef,
		VerifyLocal: true,
	}
	if remoteURL, err := git.RemoteURL(ctx, d.Runner, repoRoot, "origin"); err != nil {
		return nil, "", err
	} else if strings.TrimSpace(remoteURL) != "" {
		verifyCtx.RemoteURL = remoteURL
		verifyCtx.VerifyHosting = true
		verifyCtx.HostingProvider = hosting.DetectProvider(remoteURL)
	}

	paths := resolveListPaths(ctx, d, repoRoot)
	out := make([]cleanupCandidate, 0, len(wts))
	hostingNote := ""
	for _, wt := range wts {
		info, err := verifyWorktreeWithContext(ctx, d, verifyCtx, wt)
		if err != nil {
			return nil, "", err
		}
		signals := deriveListSignals(wt, info, paths)
		if hostingNote == "" && info != nil {
			switch info.HostingReason {
			case "gh-auth-unavailable":
				hostingNote = "note: hosting verify skipped (gh not found on PATH / WT_GH_BIN, or not authenticated)"
			case "unsupported-provider":
				if info.HostingProvider != "" && info.HostingProvider != string(hosting.ProviderUnknown) {
					hostingNote = fmt.Sprintf("note: hosting verify skipped (provider not implemented: %s)", info.HostingProvider)
				}
			}
		}
		out = append(out, cleanupCandidate{
			Worktree: wt,
			Info:     info,
			Signals:  signals,
			Reason:   cleanupReason(wt, info, signals),
		})
	}

	return out, hostingNote, nil
}

func runCleanup(cmd *cobra.Command, d *deps, repoRoot string, candidates []cleanupCandidate, apply bool) ([]cleanupItem, error) {
	items := make([]cleanupItem, 0, len(candidates))
	for _, candidate := range candidates {
		items = append(items, cleanupItemForCandidate(candidate, false))
	}
	if !apply {
		return items, nil
	}

	hasPrune := false
	for _, candidate := range candidates {
		if candidate.Signals.RecommendedAction == "prune" {
			hasPrune = true
			break
		}
	}

	remainingAfterPrune := map[string]struct{}{}
	if hasPrune {
		if err := git.WorktreePrune(cmd.Context(), d.Runner, repoRoot); err != nil {
			return nil, err
		}
		afterPrune, err := git.WorktreeList(cmd.Context(), d.Runner, repoRoot)
		if err != nil {
			return nil, err
		}
		for _, wt := range afterPrune {
			remainingAfterPrune[wt.Path] = struct{}{}
		}
	}

	for i, candidate := range candidates {
		switch candidate.Signals.RecommendedAction {
		case "prune":
			items[i].Applied = true
			if _, ok := remainingAfterPrune[candidate.Worktree.Path]; ok {
				items[i].Action = "kept"
			} else {
				items[i].Action = "pruned"
				items[i].Removed = true
			}
		case "remove":
			if !candidate.Signals.SafeToRemove {
				continue
			}
			if err := git.WorktreeRemove(cmd.Context(), d.Runner, repoRoot, candidate.Worktree.Path, true); err != nil {
				return nil, err
			}
			items[i].Action = "removed"
			items[i].Applied = true
			items[i].Removed = true
		}
	}

	return items, nil
}

func cleanupItemForCandidate(candidate cleanupCandidate, applied bool) cleanupItem {
	item := cleanupItem{
		Path:              candidate.Worktree.Path,
		Branch:            strings.TrimPrefix(candidate.Worktree.Branch, "refs/heads/"),
		RecommendedAction: candidate.Signals.RecommendedAction,
		Action:            cleanupPreviewAction(candidate.Signals.RecommendedAction),
		Applied:           applied,
		Removed:           false,
		Reason:            candidate.Reason,
		SafeToRemove:      candidate.Signals.SafeToRemove,
	}
	if candidate.Info != nil {
		item.MergedIntoBase = candidate.Info.MergedIntoBase
		item.BaseRef = candidate.Info.BaseRef
		item.MergedViaHosting = candidate.Info.MergedViaHosting
		item.HostingProvider = candidate.Info.HostingProvider
		item.HostingKind = candidate.Info.HostingKind
	}
	return item
}

func cleanupPreviewAction(recommendedAction string) string {
	switch recommendedAction {
	case "prune":
		return "would-prune"
	case "remove":
		return "would-remove"
	default:
		return "skip"
	}
}

func formatCleanupLine(item cleanupItem) string {
	line := fmt.Sprintf("%s  %s", item.Action, item.Path)
	if item.Branch != "" {
		line += fmt.Sprintf("  (%s)", item.Branch)
	}
	if item.Reason != "" {
		line += fmt.Sprintf("  [%s]", item.Reason)
	}
	return line
}

func cleanupReason(wt worktree.Worktree, info *verifyInfo, signals listSignals) string {
	switch signals.RecommendedAction {
	case "prune":
		if wt.PruneReason != "" {
			return wt.PruneReason
		}
		return "prunable"
	case "remove":
		reasons := make([]string, 0, 2)
		if info != nil && info.MergedIntoBase != nil && *info.MergedIntoBase {
			reasons = append(reasons, "merged:"+cleanupShortRef(info.BaseRef))
		}
		if info != nil && info.MergedViaHosting != nil && *info.MergedViaHosting {
			reason := "merged-hosting"
			if info.HostingProvider != "" {
				reason += ":" + info.HostingProvider
			}
			if info.HostingNumber != nil {
				reason += fmt.Sprintf("#%d", *info.HostingNumber)
			}
			reasons = append(reasons, reason)
		}
		if len(reasons) == 0 {
			return "safe-remove"
		}
		return strings.Join(reasons, ", ")
	default:
		return cleanupSkipReason(wt, info, signals)
	}
}

func cleanupSkipReason(wt worktree.Worktree, info *verifyInfo, signals listSignals) string {
	switch {
	case signals.Current:
		return "current"
	case signals.Primary:
		return "primary"
	case wt.Detached:
		return "detached"
	case wt.Locked:
		return "locked"
	}

	pathExists, dotGitExists := worktreePathStatus(wt.Path)
	if info != nil {
		pathExists = info.PathExists
		dotGitExists = info.DotGitExists
	}

	switch {
	case !pathExists:
		return "missing-path"
	case !dotGitExists:
		return "missing-git"
	case wt.Prunable:
		return "prunable"
	default:
		return "not-recommended"
	}
}

func cleanupShortRef(ref string) string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "refs/remotes/")
	ref = strings.TrimPrefix(ref, "origin/")
	ref = strings.TrimPrefix(ref, "upstream/")
	return ref
}
