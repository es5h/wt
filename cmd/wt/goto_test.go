package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"wt/internal/runner"
)

func TestGoto_PrintsOnlyPath(t *testing.T) {
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

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"goto", "feature-x"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.wt/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestGoto_JSON(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
`) + "\n"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"goto", "feature-x", "--json"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}
	if got.Path != "/repo/.wt/feature-x" || got.Branch != "feature-x" {
		t.Fatalf("unexpected json: %#v", got)
	}
}

func TestGotoCompletion_SuggestsWorktreeBranches(t *testing.T) {
	t.Setenv("WT_GOTO_COMPLETE_REMOTE", "0")

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

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
		},
	}

	cmd := newGotoCmd()
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	got, dir := cmd.ValidArgsFunction(cmd, []string{}, "")
	if dir != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want %v", dir, cobra.ShellCompDirectiveNoFileComp)
	}

	want := []string{"feature-x", "main"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("completions = %q, want %q", got, want)
	}
}

func TestGotoCompletion_FiltersByPrefix(t *testing.T) {
	t.Setenv("WT_GOTO_COMPLETE_REMOTE", "0")

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

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
		},
	}

	cmd := newGotoCmd()
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	got, _ := cmd.ValidArgsFunction(cmd, []string{}, "f")
	if strings.Join(got, ",") != "feature-x" {
		t.Fatalf("completions = %q, want %q", got, []string{"feature-x"})
	}
}

func TestGotoCompletion_RemoteOptIn(t *testing.T) {
	t.Setenv("WT_GOTO_COMPLETE_REMOTE", "1")

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"for-each-ref", "--format=%(refname:strip=3)", "refs/remotes/origin"},
				res: runner.Result{
					Stdout:   []byte("feature-y\nHEAD\n"),
					ExitCode: 0,
				},
			},
		},
	}

	cmd := newGotoCmd()
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	got, _ := cmd.ValidArgsFunction(cmd, []string{}, "f")
	if strings.Join(got, ",") != "feature-y" {
		t.Fatalf("completions = %q, want %q", got, []string{"feature-y"})
	}
}

func TestGoto_CreateWhenMissing(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"config", "--local", "--get", "wt.root"},
				res: runner.Result{
					ExitCode: 1,
				},
				err: errors.New("exit 1"),
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"rev-parse", "--verify", "--quiet", "refs/heads/feature-x^{commit}"},
				res: runner.Result{
					ExitCode: 1,
				},
				err: errors.New("exit 1"),
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"},
				res: runner.Result{
					Stdout:   []byte("refs/remotes/origin/main\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"rev-parse", "--verify", "--quiet", "refs/remotes/origin/feature-x^{commit}"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"rev-parse", "--verify", "--quiet", "origin/feature-x^{commit}"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "add", "-b", "feature-x", "/repo/.wt/feature-x", "origin/feature-x"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"goto", "feature-x", "--create"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.wt/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestGoto_CreateUsesRootFlag(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--show-toplevel"},
				res: runner.Result{
					Stdout:   []byte(repo + "\n"),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res: runner.Result{
					Stdout:   []byte(porcelain),
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"rev-parse", "--verify", "--quiet", "refs/heads/feature-x^{commit}"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "add", "/alt-root/feature-x", "feature-x"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"goto", "feature-x", "--create", "--root", "/alt-root"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/alt-root/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}
