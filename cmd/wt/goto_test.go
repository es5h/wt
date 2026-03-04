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
	"wt/internal/tui/picker"
	"wt/internal/worktree"
)

func TestPath_PrintsOnlyPath(t *testing.T) {
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
	root.SetArgs([]string{"path", "feature-x"})
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

func TestPath_JSON(t *testing.T) {
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
	root.SetArgs([]string{"path", "feature-x", "--json"})
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

func TestPathCompletion_SuggestsWorktreeBranches(t *testing.T) {
	t.Setenv("WT_PATH_COMPLETE_REMOTE", "0")

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

	cmd := newPathCmd()
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

func TestPathCompletion_FiltersByPrefix(t *testing.T) {
	t.Setenv("WT_PATH_COMPLETE_REMOTE", "0")

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

	cmd := newPathCmd()
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	got, _ := cmd.ValidArgsFunction(cmd, []string{}, "f")
	if strings.Join(got, ",") != "feature-x" {
		t.Fatalf("completions = %q, want %q", got, []string{"feature-x"})
	}
}

func TestPathCompletion_RemoteOptIn(t *testing.T) {
	t.Setenv("WT_PATH_COMPLETE_REMOTE", "1")

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

	cmd := newPathCmd()
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	got, _ := cmd.ValidArgsFunction(cmd, []string{}, "f")
	if strings.Join(got, ",") != "feature-y" {
		t.Fatalf("completions = %q, want %q", got, []string{"feature-y"})
	}
}

func TestPath_CreateRemoteBranchOnly(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
	root.SetArgs([]string{"path", "feature-x", "--create"})
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

func TestPath_CreateAttachesExistingLocalBranchWithoutNewBranch(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
	root.SetArgs([]string{"path", "feature-x", "--create"})
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

func TestPath_CreateNoMatchUsesDefaultBase(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
				args:    []string{"rev-parse", "--verify", "--quiet", "refs/heads/brand-new^{commit}"},
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
				args:    []string{"rev-parse", "--verify", "--quiet", "refs/remotes/origin/brand-new^{commit}"},
				res: runner.Result{
					ExitCode: 1,
				},
				err: errors.New("exit 1"),
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"rev-parse", "--verify", "--quiet", "origin/main^{commit}"},
				res: runner.Result{
					ExitCode: 0,
				},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "add", "-b", "brand-new", "/repo/.wt/brand-new", "origin/main"},
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
	root.SetArgs([]string{"path", "brand-new", "--create"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.wt/brand-new\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestPath_CreateFailsWhenPrunableEntryExists(t *testing.T) {
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
	root.SetArgs([]string{"path", "feature-x", "--create"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(err.Error(), "registered worktree entry is prunable") {
		t.Fatalf("error = %v, want prunable guidance", err)
	}
	if !strings.Contains(err.Error(), "wt prune --apply") {
		t.Fatalf("error = %v, want prune guidance", err)
	}
}

func TestPath_CreateUsesRootFlag(t *testing.T) {
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
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte("/repo/.git\n"),
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
	root.SetArgs([]string{"path", "feature-x", "--create", "--root", "/alt-root"})
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

func TestPathTUI_RequiresTTY(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{"path", "--tui"})
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
	if !strings.Contains(err.Error(), "--tui requires a TTY") {
		t.Fatalf("err = %q, want TTY guidance", err.Error())
	}
}

func TestPathTUI_WithoutQuery_UsesWholeWorktreeList(t *testing.T) {
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
				res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"path", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner:    r,
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PickWorktree: func(_ *cobra.Command, wts []worktree.Worktree, initialFilter string) (worktree.Worktree, error) {
			if initialFilter != "" {
				t.Fatalf("initialFilter = %q, want empty", initialFilter)
			}
			if len(wts) != 2 {
				t.Fatalf("len(wts) = %d, want 2", len(wts))
			}
			return wts[0], nil
		},
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestPathTUI_UsesPickerSelection(t *testing.T) {
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

worktree /repo/.wt/feature-y
HEAD fedcbafedcbafedcbafedcbafedcbafedcbafedc
branch refs/heads/feature-y
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
	root.SetArgs([]string{"path", "feature", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner:    r,
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PickWorktree: func(_ *cobra.Command, wts []worktree.Worktree, initialFilter string) (worktree.Worktree, error) {
			if initialFilter != "feature" {
				t.Fatalf("initialFilter = %q, want %q", initialFilter, "feature")
			}
			if len(wts) != 2 {
				t.Fatalf("len(wts) = %d, want 2", len(wts))
			}
			if wts[0].Path != "/repo/.wt/feature-x" || wts[1].Path != "/repo/.wt/feature-y" {
				t.Fatalf("picker candidates = %#v, want only matched feature worktrees", wts)
			}
			return wts[1], nil
		},
	}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "/repo/.wt/feature-y\n" {
		t.Fatalf("stdout = %q, want only path", stdout.String())
	}
}

func TestPathTUI_WithSingleQueryMatch_SkipsPicker(t *testing.T) {
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
				res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
			},
			{
				workDir: repo,
				name:    "git",
				args:    []string{"worktree", "list", "--porcelain"},
				res:     runner.Result{Stdout: []byte(porcelain), ExitCode: 0},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"path", "feature-x", "--tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner:    r,
		Cwd:       cwd,
		CanUseTUI: func() bool { return true },
		PickWorktree: func(_ *cobra.Command, _ []worktree.Worktree, _ string) (worktree.Worktree, error) {
			t.Fatal("picker should not be called for a single match")
			return worktree.Worktree{}, nil
		},
	}))

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

func TestPathTUI_CancelReturnsExit130(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
`) + "\n"

	root := newRootCmd()
	root.SetArgs([]string{"path", "--tui"})
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
		PickWorktree: func(_ *cobra.Command, _ []worktree.Worktree, _ string) (worktree.Worktree, error) {
			return worktree.Worktree{}, picker.ErrCancelled
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
	if err.Error() != "wt path: selection cancelled" {
		t.Fatalf("err = %q, want cancellation message", err.Error())
	}
}

func TestPathRejectsCreateOnlyFlagsWithoutCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "path flag",
			args: []string{"path", "feature-x", "--path", "/tmp/feature-x"},
			want: "wt path: --path requires --create",
		},
		{
			name: "root flag",
			args: []string{"path", "feature-x", "--root", "/tmp"},
			want: "wt path: --root requires --create",
		},
		{
			name: "from flag",
			args: []string{"path", "feature-x", "--from", "origin/main"},
			want: "wt path: --from requires --create",
		},
		{
			name: "dry-run flag",
			args: []string{"path", "feature-x", "--dry-run"},
			want: "wt path: --dry-run requires --create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := newRootCmd()
			root.SetArgs(tt.args)
			root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{}))

			err := root.Execute()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var exitErr *exitError
			if !errors.As(err, &exitErr) || exitErr.Code != 2 {
				t.Fatalf("err = %#v, want exitError code 2", err)
			}
			if err.Error() != tt.want {
				t.Fatalf("err = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestPathRejectsNoTUIWithoutQuery(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{"path", "--no-tui"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("err = %#v, want exitError code 2", err)
	}
	if err.Error() != "wt path: query is required when --no-tui is set" {
		t.Fatalf("err = %q, want no-tui guidance", err.Error())
	}
}
