package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestRun_ForwardsStdoutStderr(t *testing.T) {
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
			{
				workDir: "/repo/.wt/feature-x",
				name:    "go",
				args:    []string{"test", "./..."},
				res: runner.Result{
					Stdout:   []byte("ok\t./...\n"),
					Stderr:   []byte("warning\n"),
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"run", "feature-x", "--", "go", "test", "./..."})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.String() != "ok\t./...\n" {
		t.Fatalf("stdout = %q, want forwarded child stdout", stdout.String())
	}
	if stderr.String() != "warning\n" {
		t.Fatalf("stderr = %q, want forwarded child stderr", stderr.String())
	}
}

func TestRun_JSONPreservesExitCode(t *testing.T) {
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
			{
				workDir: "/repo/.wt/feature-x",
				name:    "go",
				args:    []string{"test", "./..."},
				res: runner.Result{
					Stdout:   []byte("ok\t./...\n"),
					Stderr:   []byte("warning\n"),
					ExitCode: 23,
				},
				err: errors.New("exit 23"),
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"run", "feature-x", "--json", "--", "go", "test", "./..."})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	err := root.Execute()
	var exitErr *exitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Execute() error = %v, want exitError", err)
	}
	if exitErr.Code != 23 {
		t.Fatalf("exit code = %d, want 23", exitErr.Code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty in --json mode", stderr.String())
	}

	var got struct {
		Path     string   `json:"path"`
		Command  []string `json:"command"`
		ExitCode int      `json:"exitCode"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}
	if got.Path != "/repo/.wt/feature-x" {
		t.Fatalf("path = %q, want %q", got.Path, "/repo/.wt/feature-x")
	}
	if strings.Join(got.Command, "\x00") != strings.Join([]string{"go", "test", "./..."}, "\x00") {
		t.Fatalf("command = %#v, want %#v", got.Command, []string{"go", "test", "./..."})
	}
	if got.ExitCode != 23 {
		t.Fatalf("exitCode = %d, want 23", got.ExitCode)
	}
}

func TestRun_AmbiguousQuery_HintsPathTUI(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"
	porcelain := strings.TrimSpace(`
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
	root.SetArgs([]string{"run", "feature", "--", "pwd"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("err = %#v, want exitError code 1", err)
	}
	if !strings.Contains(err.Error(), "wt run: 2 matches for \"feature\"") {
		t.Fatalf("err = %q, want ambiguous match summary", err.Error())
	}
	if !strings.Contains(err.Error(), "resolve the path first with `wt path <query> --tui`") {
		t.Fatalf("err = %q, want wt path --tui guidance", err.Error())
	}
}

func TestRun_UsageErrorExitCode2(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	root.SetArgs([]string{"run", "feature-x"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("err = %#v, want exitError code 2", err)
	}
	if !strings.Contains(err.Error(), "wt run: requires <query> and <cmd...>") {
		t.Fatalf("err = %q, want usage message", err.Error())
	}
}
