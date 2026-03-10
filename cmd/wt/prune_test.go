package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/runner"
	"github.com/es5h/wt/internal/tui/picker"
)

func TestPrune_Preview(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "would-prune  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want preview line", stdout.String())
	}
}

func TestPrune_Apply(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	before := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"
	after := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune", "--apply"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(before), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "prune", "--expire", "now"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(after), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "pruned  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want pruned line", stdout.String())
	}
}

func TestPrune_JSON(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune", "--json"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0]["action"] != "preview" {
		t.Fatalf("action = %#v, want preview", got[0]["action"])
	}
	if got[0]["removed"] != false {
		t.Fatalf("removed = %#v, want false", got[0]["removed"])
	}
}

func TestPruneTUI_RequiresTTY(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{"prune", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		CanUseTUI: func() bool { return false },
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("err = %#v, want exitError code 2", err)
	}
	if err.Error() != "wt prune: --tui requires a TTY on stdin and stderr" {
		t.Fatalf("err = %q, want TTY guidance", err.Error())
	}
}

func TestPruneTUI_NoCandidates_SkipsPreview(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PreviewPrune: func(_ *cobra.Command, items []pruneCandidate, apply bool) error {
			t.Fatalf("PreviewPrune should not be called, got %d items apply=%v", len(items), apply)
			return nil
		},
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPruneTUI_PreviewShowsOnlyPrunableEntries(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location

worktree /repo/.wt/feature-y
HEAD fedcbafedcbafedcbafedcbafedcbafedcbafedc
branch refs/heads/feature-y
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PreviewPrune: func(_ *cobra.Command, items []pruneCandidate, apply bool) error {
			if apply {
				t.Fatal("apply = true, want false")
			}
			if len(items) != 1 {
				t.Fatalf("len(items) = %d, want 1", len(items))
			}
			if items[0].Path != "/repo/.wt/feature-x" {
				t.Fatalf("items[0].Path = %q, want feature-x", items[0].Path)
			}
			if items[0].PruneReason != "gitdir file points to non-existent location" {
				t.Fatalf("items[0].PruneReason = %q", items[0].PruneReason)
			}
			return nil
		},
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "would-prune  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want preview line", stdout.String())
	}
	if strings.Contains(stdout.String(), "feature-y") {
		t.Fatalf("stdout = %q, want only prunable entries", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPruneTUI_ApplyConfirm(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	before := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"
	after := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetIn(strings.NewReader("yes\n"))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"prune", "--tui", "--apply"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(before), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "prune", "--expire", "now"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(after), ExitCode: 0},
				},
			},
		},
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PreviewPrune: func(_ *cobra.Command, items []pruneCandidate, apply bool) error {
			if !apply {
				t.Fatal("apply = false, want true")
			}
			if len(items) != 1 || items[0].Path != "/repo/.wt/feature-x" {
				t.Fatalf("items = %#v, want single feature-x candidate", items)
			}
			return nil
		},
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "Prune 1 stale worktree entry with git worktree prune --expire now? [y/N]") {
		t.Fatalf("stderr = %q, want confirmation prompt", stderr.String())
	}
	if !strings.Contains(stdout.String(), "pruned  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want pruned line", stdout.String())
	}
}

func TestPruneTUI_ConfirmCancelAborts(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"

	root := newRootCmd()
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"prune", "--tui", "--apply"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PreviewPrune: func(_ *cobra.Command, _ []pruneCandidate, _ bool) error {
			return nil
		},
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("err = %#v, want exitError code 1", err)
	}
	if err.Error() != "wt prune: aborted" {
		t.Fatalf("err = %q, want aborted", err.Error())
	}
}

func TestPruneTUI_PreviewCancelReturnsExit130(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"

	root := newRootCmd()
	root.SetArgs([]string{"prune", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PreviewPrune: func(_ *cobra.Command, _ []pruneCandidate, _ bool) error {
			return picker.ErrCancelled
		},
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 130 {
		t.Fatalf("err = %#v, want exitError code 130", err)
	}
	if err.Error() != "wt prune: preview cancelled" {
		t.Fatalf("err = %q, want preview cancelled", err.Error())
	}
}

func TestPruneTUI_RejectsJSON(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{"prune", "--tui", "--json"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("err = %#v, want exitError code 2", err)
	}
	if err.Error() != "wt prune: --tui cannot be combined with --json" {
		t.Fatalf("err = %q, want json/tui guidance", err.Error())
	}
}

func TestPreviewPruneWithTUIBuildsPickerRows(t *testing.T) {
	t.Parallel()

	items := buildPrunePickerItems([]pruneCandidate{
		{
			Path:        "/repo/.wt/feature-x",
			Branch:      "feature-x",
			PruneReason: "gitdir file points to non-existent location",
		},
		{
			Path:        "/repo/.wt/stale-only",
			PruneReason: "missing worktree path",
		},
	})

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Label != "feature-x" {
		t.Fatalf("items[0].Label = %q, want feature-x", items[0].Label)
	}
	if items[0].Detail != "/repo/.wt/feature-x" {
		t.Fatalf("items[0].Detail = %q", items[0].Detail)
	}
	if !strings.Contains(items[0].Meta, "prunable") || !strings.Contains(items[0].Meta, "gitdir file points to non-existent location") {
		t.Fatalf("items[0].Meta = %q, want prunable reason metadata", items[0].Meta)
	}
	if items[1].Label != "stale-only" {
		t.Fatalf("items[1].Label = %q, want basename fallback", items[1].Label)
	}
	if !strings.Contains(items[1].FilterText, "missing worktree path") {
		t.Fatalf("items[1].FilterText = %q, want prune reason included", items[1].FilterText)
	}
}
