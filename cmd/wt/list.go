package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"wt/internal/git"
	"wt/internal/hosting"
	"wt/internal/worktree"
)

func newListCmd() *cobra.Command {
	var jsonOut bool
	var porcelain bool
	var verify bool
	var verifyHosting bool
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
			if porcelain && verifyHosting {
				return usageError(fmt.Errorf("wt list: --porcelain cannot be combined with --verify-hosting"))
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
			if verify || verifyHosting {
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
				if verifyHosting {
					remoteURL, err := git.RemoteURL(ctx, d.Runner, repoRoot, "origin")
					if err != nil {
						return err
					}
					verifyCtx.RemoteURL = remoteURL
					verifyCtx.VerifyHosting = true
					verifyCtx.HostingProvider = hosting.DetectProvider(remoteURL)
				}
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(toJSONWorktrees(cmd, d, wts, verifyCtx))
			}

			hostingNote := formatHostingVerifyNote(wts, d, verifyCtx)
			if hostingNote != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), hostingNote)
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
	cmd.Flags().BoolVar(&verifyHosting, "verify-hosting", false, "opt-in hosting merge verification (GitHub via gh only; GitLab reserved)")
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

	Verify *jsonVerifyFields `json:"-"`
}

type jsonVerifyFields struct {
	PathExists       bool
	DotGitExists     bool
	Valid            bool
	MergedIntoBase   *bool
	BaseRef          string
	HostingProvider  string
	MergedViaHosting *bool
	HostingReason    string
}

type listVerifyContext struct {
	RepoRoot        string
	BaseRef         string
	RemoteURL       string
	VerifyHosting   bool
	HostingProvider hosting.Provider
}

type verifyInfo struct {
	PathExists       bool
	DotGitExists     bool
	Valid            bool
	MergedIntoBase   *bool
	BaseRef          string
	HostingProvider  string
	MergedViaHosting *bool
	HostingReason    string
}

func (jwt jsonWorktree) MarshalJSON() ([]byte, error) {
	type baseJSONWorktree struct {
		Path   string `json:"path"`
		HEAD   string `json:"head"`
		Branch string `json:"branch"`

		Detached bool `json:"detached"`
		Locked   bool `json:"locked"`
		Prunable bool `json:"prunable"`

		LockReason  string `json:"lockReason,omitempty"`
		PruneReason string `json:"pruneReason,omitempty"`

		PathExists       *bool  `json:"pathExists,omitempty"`
		DotGitExists     *bool  `json:"dotGitExists,omitempty"`
		Valid            *bool  `json:"valid,omitempty"`
		MergedIntoBase   *bool  `json:"mergedIntoBase,omitempty"`
		BaseRef          string `json:"baseRef,omitempty"`
		HostingProvider  string `json:"hostingProvider,omitempty"`
		MergedViaHosting *bool  `json:"mergedViaHosting,omitempty"`
		HostingReason    string `json:"hostingReason,omitempty"`
	}

	out := baseJSONWorktree{
		Path:        jwt.Path,
		HEAD:        jwt.HEAD,
		Branch:      jwt.Branch,
		Detached:    jwt.Detached,
		Locked:      jwt.Locked,
		Prunable:    jwt.Prunable,
		LockReason:  jwt.LockReason,
		PruneReason: jwt.PruneReason,
	}
	if jwt.Verify != nil {
		out.PathExists = &jwt.Verify.PathExists
		out.DotGitExists = &jwt.Verify.DotGitExists
		out.Valid = &jwt.Verify.Valid
		out.MergedIntoBase = jwt.Verify.MergedIntoBase
		out.BaseRef = jwt.Verify.BaseRef
		out.HostingProvider = jwt.Verify.HostingProvider
		out.MergedViaHosting = jwt.Verify.MergedViaHosting
		out.HostingReason = jwt.Verify.HostingReason
	}

	if jwt.Verify != nil && (jwt.Verify.MergedIntoBase == nil || jwt.Verify.HostingProvider != "") {
		outMap := map[string]any{
			"path":     jwt.Path,
			"head":     jwt.HEAD,
			"branch":   jwt.Branch,
			"detached": jwt.Detached,
			"locked":   jwt.Locked,
			"prunable": jwt.Prunable,
		}
		if jwt.LockReason != "" {
			outMap["lockReason"] = jwt.LockReason
		}
		if jwt.PruneReason != "" {
			outMap["pruneReason"] = jwt.PruneReason
		}
		outMap["pathExists"] = jwt.Verify.PathExists
		outMap["dotGitExists"] = jwt.Verify.DotGitExists
		outMap["valid"] = jwt.Verify.Valid
		outMap["mergedIntoBase"] = jwt.Verify.MergedIntoBase
		outMap["baseRef"] = jwt.Verify.BaseRef
		if jwt.Verify.HostingProvider != "" {
			outMap["hostingProvider"] = jwt.Verify.HostingProvider
			outMap["mergedViaHosting"] = jwt.Verify.MergedViaHosting
			if jwt.Verify.HostingReason != "" {
				outMap["hostingReason"] = jwt.Verify.HostingReason
			}
		}
		return json.Marshal(outMap)
	}

	return json.Marshal(out)
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
				jwt.Verify = &jsonVerifyFields{
					PathExists:       info.PathExists,
					DotGitExists:     info.DotGitExists,
					Valid:            info.Valid,
					MergedIntoBase:   info.MergedIntoBase,
					BaseRef:          info.BaseRef,
					HostingProvider:  info.HostingProvider,
					MergedViaHosting: info.MergedViaHosting,
					HostingReason:    info.HostingReason,
				}
			}
		}
		out = append(out, jwt)
	}
	return out
}

func verifyWorktree(cmd *cobra.Command, d *deps, verifyCtx *listVerifyContext, wt worktree.Worktree) (*verifyInfo, error) {
	var ctx context.Context
	if cmd != nil {
		ctx = cmd.Context()
	}
	return verifyWorktreeWithContext(ctx, d, verifyCtx, wt)
}

func verifyWorktreeWithContext(ctx context.Context, d *deps, verifyCtx *listVerifyContext, wt worktree.Worktree) (*verifyInfo, error) {
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
		isMerged, err := git.IsAncestor(ctx, d.Runner, verifyCtx.RepoRoot, wt.Branch, verifyCtx.BaseRef)
		if err != nil {
			return nil, err
		}
		merged = &isMerged
	}

	var hostingMerged *bool
	hostingProvider := ""
	hostingReason := ""
	if verifyCtx.VerifyHosting {
		hostingProvider = string(verifyCtx.HostingProvider)
		if wt.Branch == "" || wt.Detached {
			hostingReason = "no-branch"
		} else {
			result, err := hosting.VerifyMerged(ctx, d.Runner, verifyCtx.RepoRoot, verifyCtx.HostingProvider, strings.TrimPrefix(wt.Branch, "refs/heads/"), verifyCtx.BaseRef)
			if err != nil {
				return nil, err
			}
			hostingProvider = string(result.Provider)
			hostingMerged = result.Merged
			hostingReason = result.Reason
		}
	}

	return &verifyInfo{
		PathExists:       pathExists,
		DotGitExists:     dotGitExists,
		Valid:            valid,
		MergedIntoBase:   merged,
		BaseRef:          verifyCtx.BaseRef,
		HostingProvider:  hostingProvider,
		MergedViaHosting: hostingMerged,
		HostingReason:    hostingReason,
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
		if info.MergedViaHosting != nil && *info.MergedViaHosting {
			flags = append(flags, "merged-pr")
		}
	}

	if len(flags) == 0 {
		return fmt.Sprintf("%s  %s  %s  %s", base, branch, head, wt.Path)
	}
	return fmt.Sprintf("%s  %s  %s  %s  [%s]", base, branch, head, wt.Path, strings.Join(flags, ","))
}

func formatHostingVerifyNote(wts []worktree.Worktree, d *deps, verifyCtx *listVerifyContext) string {
	if verifyCtx == nil || !verifyCtx.VerifyHosting || d == nil {
		return ""
	}

	for _, wt := range wts {
		info, err := verifyWorktreeWithContext(context.Background(), d, verifyCtx, wt)
		if err != nil || info == nil {
			continue
		}
		switch info.HostingReason {
		case "gh-auth-unavailable":
			return "note: hosting verify skipped (gh not found on PATH / WT_GH_BIN, or not authenticated)"
		case "unsupported-provider":
			if info.HostingProvider != "" && info.HostingProvider != string(hosting.ProviderUnknown) {
				return fmt.Sprintf("note: hosting verify skipped (provider not implemented: %s)", info.HostingProvider)
			}
		}
	}

	return ""
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
