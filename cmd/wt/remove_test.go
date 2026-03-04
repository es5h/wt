package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"wt/internal/runner"
)

func TestRemove_DryRun(t *testing.T) {
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
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"remove", "feature-x", "--dry-run"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "would-remove  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want preview line", stdout.String())
	}
}

func TestRemove_Force(t *testing.T) {
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
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"remove", "feature-x", "--force"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "remove", "--force", "/repo/.wt/feature-x"},
					res:     runner.Result{ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want removed line", stdout.String())
	}
}

func TestRemove_InteractiveConfirmYes(t *testing.T) {
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
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetIn(strings.NewReader("y\n"))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"remove", "feature-x"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "remove", "--force", "/repo/.wt/feature-x"},
					res:     runner.Result{ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return true },
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "Remove worktree /repo/.wt/feature-x (feature-x)? [y/N] ") {
		t.Fatalf("stderr = %q, want prompt", stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed  /repo/.wt/feature-x  (feature-x)") {
		t.Fatalf("stdout = %q, want removed line", stdout.String())
	}
}

func TestRemove_InteractiveConfirmNo(t *testing.T) {
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
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetIn(strings.NewReader("n\n"))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"remove", "feature-x"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return true },
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "wt remove: aborted" {
		t.Fatalf("err = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Remove worktree /repo/.wt/feature-x (feature-x)? [y/N] ") {
		t.Fatalf("stderr = %q, want prompt", stderr.String())
	}
}

func TestRemove_JSON(t *testing.T) {
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
`) + "\n"

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"remove", "feature-x", "--dry-run", "--json"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got["action"] != "preview" {
		t.Fatalf("action = %#v, want preview", got["action"])
	}
	if got["removed"] != false {
		t.Fatalf("removed = %#v, want false", got["removed"])
	}
}

func TestRemove_RefusesWithoutFlagOnNonTTY(t *testing.T) {
	t.Parallel()

	cmd := newRemoveCmd()
	cmd.SetArgs([]string{"feature-x"})
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner:        &fakeRunner{t: t},
		Cwd:           "/cwd",
		IsInteractive: func() bool { return false },
	}))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "wt remove: requires --dry-run or --force" {
		t.Fatalf("err = %v", err)
	}
}

func TestRemove_RefusesCurrentWorktree(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	root := newRootCmd()
	root.SetArgs([]string{"remove", "main", "--dry-run"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "wt remove: cannot remove current worktree: /repo" {
		t.Fatalf("err = %v", err)
	}
}

func TestRemove_RefusesPrimaryWorktree(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo/.wt/current"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main

worktree /repo/.wt/current
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/current
`) + "\n"

	root := newRootCmd()
	root.SetArgs([]string{"remove", "main", "--dry-run"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "wt remove: cannot remove primary worktree: /repo" {
		t.Fatalf("err = %v", err)
	}
}

func TestRemove_RefusesPrunableTarget(t *testing.T) {
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
	root.SetArgs([]string{"remove", "feature-x", "--dry-run"})
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte("/repo/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"worktree", "list", "--porcelain"},
					res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
				},
			},
		},
		Cwd:           cwd,
		IsInteractive: func() bool { return false },
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "use 'wt prune --apply'") {
		t.Fatalf("err = %v", err)
	}
}
