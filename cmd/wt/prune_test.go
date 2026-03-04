package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"wt/internal/runner"
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
