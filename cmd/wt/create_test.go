package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestCreate_UsesRemoteIfExists(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
	root.SetArgs([]string{"create", "feature-x"})
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

func TestCreate_LocalBranchExistsButNoLiveWorktree(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "add", "/repo/.wt/feature-x", "feature-x"},
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
	root.SetArgs([]string{"create", "feature-x"})
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

func TestCreate_FailsWhenPrunableEntryExists(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
	root.SetArgs([]string{"create", "feature-x"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(err.Error(), "registered worktree entry is prunable") {
		t.Fatalf("error = %v, want prunable guidance", err)
	}
	if !strings.Contains(err.Error(), "wt prune --apply") {
		t.Fatalf("error = %v, want prune guidance", err)
	}
}

func TestCreate_UsesRootFlagOverEnvAndConfig(t *testing.T) {
	t.Setenv("WT_ROOT", "/env-root")

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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
				args:    []string{"worktree", "add", "/flag-root/feature-x", "feature-x"},
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
	root.SetArgs([]string{"create", "feature-x", "--root", "/flag-root"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/flag-root/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestCreate_UsesEnvRootWhenFlagMissing(t *testing.T) {
	t.Setenv("WT_ROOT", ".trees")

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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
				args:    []string{"worktree", "add", "/repo/.trees/feature-x", "feature-x"},
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
	root.SetArgs([]string{"create", "feature-x"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.trees/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestCreate_UsesRepoLocalConfigRootWhenEnvMissing(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
					Stdout:   []byte(".trees\n"),
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
				args:    []string{"worktree", "add", "/repo/.trees/feature-x", "feature-x"},
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
	root.SetArgs([]string{"create", "feature-x"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.trees/feature-x\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}
