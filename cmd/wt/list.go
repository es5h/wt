package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/worktree"
)

func newListCmd() *cobra.Command {
	var jsonOut bool
	var porcelain bool
	var verify bool
	var baseRef string

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List worktrees in current repo",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOut && porcelain {
				return usageError(fmt.Errorf("wt list: only one of --json or --porcelain can be specified"))
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

			if porcelain {
				out, err := git.WorktreeListPorcelain(ctx, d.Runner, repoRoot)
				if err != nil {
					return err
				}
				_, _ = cmd.OutOrStdout().Write(out)
				return nil
			}

			wts, err := git.WorktreeList(ctx, d.Runner, repoRoot)
			if err != nil {
				return err
			}

			var verifyCtx *listVerifyContext
			if verify {
				if baseRef == "" {
					baseRef = git.DefaultBaseRef(ctx, d.Runner, repoRoot)
				}
				exists, err := git.RefExists(ctx, d.Runner, repoRoot, baseRef)
				if err != nil {
					return err
				}
				if !exists {
					return usageError(fmt.Errorf("wt list: base ref does not exist: %s", baseRef))
				}
				verifyCtx = &listVerifyContext{RepoRoot: repoRoot, BaseRef: baseRef}
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(toJSONWorktrees(cmd, d, wts, verifyCtx))
			}

			for _, wt := range wts {
				info, _ := verifyWorktree(cmd, d, verifyCtx, wt)
				fmt.Fprintln(cmd.OutOrStdout(), formatWorktreeLine(wt, info))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	cmd.Flags().BoolVar(&porcelain, "porcelain", false, "git porcelain output (for parsing)")
	cmd.Flags().BoolVar(&verify, "verify", false, "verify worktree entries (checks path and merged-to-base)")
	cmd.Flags().StringVar(&baseRef, "base", "", "base ref for --verify (default: origin/HEAD or main)")
	return cmd
}

type jsonWorktree struct {
	Path   string `json:"path"`
	HEAD   string `json:"head"`
	Branch string `json:"branch"`

	Detached bool `json:"detached"`
	Locked   bool `json:"locked"`
	Prunable bool `json:"prunable"`

	LockReason  string `json:"lockReason,omitempty"`
	PruneReason string `json:"pruneReason,omitempty"`

	PathExists     *bool  `json:"pathExists,omitempty"`
	DotGitExists   *bool  `json:"dotGitExists,omitempty"`
	Valid          *bool  `json:"valid,omitempty"`
	MergedIntoBase *bool  `json:"mergedIntoBase,omitempty"`
	BaseRef        string `json:"baseRef,omitempty"`
}

type listVerifyContext struct {
	RepoRoot string
	BaseRef  string
}

type verifyInfo struct {
	PathExists     bool
	DotGitExists   bool
	Valid          bool
	MergedIntoBase *bool
	BaseRef        string
}

func toJSONWorktrees(cmd *cobra.Command, d *deps, wts []worktree.Worktree, verifyCtx *listVerifyContext) []jsonWorktree {
	out := make([]jsonWorktree, 0, len(wts))
	for _, wt := range wts {
		jwt := jsonWorktree{
			Path:        wt.Path,
			HEAD:        wt.HEAD,
			Branch:      wt.Branch,
			Detached:    wt.Detached,
			Locked:      wt.Locked,
			Prunable:    wt.Prunable,
			LockReason:  wt.LockReason,
			PruneReason: wt.PruneReason,
		}
		if verifyCtx != nil {
			info, _ := verifyWorktree(cmd, d, verifyCtx, wt)
			if info != nil {
				jwt.PathExists = &info.PathExists
				jwt.DotGitExists = &info.DotGitExists
				jwt.Valid = &info.Valid
				jwt.MergedIntoBase = info.MergedIntoBase
				jwt.BaseRef = info.BaseRef
			}
		}
		out = append(out, jwt)
	}
	return out
}

func verifyWorktree(cmd *cobra.Command, d *deps, verifyCtx *listVerifyContext, wt worktree.Worktree) (*verifyInfo, error) {
	if verifyCtx == nil {
		return nil, nil
	}

	_, err := os.Stat(wt.Path)
	pathExists := err == nil

	_, err = os.Stat(filepath.Join(wt.Path, ".git"))
	dotGitExists := err == nil

	valid := pathExists && dotGitExists && !wt.Prunable

	var merged *bool
	if d != nil && wt.Branch != "" && !wt.Detached {
		isMerged, err := git.IsAncestor(cmd.Context(), d.Runner, verifyCtx.RepoRoot, wt.Branch, verifyCtx.BaseRef)
		if err != nil {
			return nil, err
		}
		merged = &isMerged
	}

	return &verifyInfo{
		PathExists:     pathExists,
		DotGitExists:   dotGitExists,
		Valid:          valid,
		MergedIntoBase: merged,
		BaseRef:        verifyCtx.BaseRef,
	}, nil
}

func formatWorktreeLine(wt worktree.Worktree, info *verifyInfo) string {
	head := wt.HEAD
	if len(head) > 7 {
		head = head[:7]
	}

	branch := displayBranch(wt)
	base := filepath.Base(wt.Path)

	flags := make([]string, 0, 2)
	if wt.Locked {
		flags = append(flags, "locked")
	}
	if wt.Prunable {
		flags = append(flags, "prunable")
	}

	if info != nil {
		if !info.PathExists {
			flags = append(flags, "missing-path")
		}
		if !info.DotGitExists {
			flags = append(flags, "missing-git")
		}
		if info.MergedIntoBase != nil && *info.MergedIntoBase {
			flags = append(flags, "merged")
		}
	}

	if len(flags) == 0 {
		return fmt.Sprintf("%s  %s  %s  %s", base, branch, head, wt.Path)
	}
	return fmt.Sprintf("%s  %s  %s  %s  [%s]", base, branch, head, wt.Path, strings.Join(flags, ","))
}

func displayBranch(wt worktree.Worktree) string {
	if wt.Branch != "" {
		return strings.TrimPrefix(wt.Branch, "refs/heads/")
	}
	if wt.Detached {
		return "(detached)"
	}
	return "(unknown)"
}
