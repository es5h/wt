package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"wt/internal/runner"
)

func newListCmdWithDeps(t *testing.T, d *deps) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	cmd := newListCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, d))
	return cmd, &stdout, &stderr
}

func TestList_Porcelain(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`) + "\n"

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{
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
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--porcelain"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != porcelain {
		t.Fatalf("stdout = %q, want porcelain output", stdout.String())
	}
}

func TestList_JSON(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	porcelain := strings.TrimSpace(`
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
locked reason: manually locked
`) + "\n"

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{
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
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got []jsonWorktree
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	if got[0].Path != "/repo" || got[0].Branch != "refs/heads/main" || got[0].Locked != true {
		t.Fatalf("unexpected json: %#v", got[0])
	}
	if got[0].LockReason == "" {
		t.Fatalf("expected lockReason to be present")
	}
}

func TestList_FlagsMutualExclusion(t *testing.T) {
	t.Parallel()

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{t: t},
		Cwd:    "/cwd",
	})
	cmd.SetArgs([]string{"--json", "--porcelain"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("Execute() error = nil, want non-nil")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("error = %#v, want exitError code 2", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty (cobra prints errors in main)", stderr.String())
	}
}

func TestList_Verify_Merged(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	wtPath := filepath.Join(t.TempDir(), "feature-x")
	if err := os.MkdirAll(filepath.Join(wtPath, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	porcelain := strings.TrimSpace(`
worktree `+wtPath+`
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
`) + "\n"

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res: runner.Result{
						ExitCode: 0,
					},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"merge-base", "--is-ancestor", "refs/heads/feature-x", "main"},
					res: runner.Result{
						ExitCode: 0,
					},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--verify", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "[merged]") {
		t.Fatalf("stdout = %q, want merged marker", stdout.String())
	}
}

func TestList_Verify_JSONIncludesFields(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	wtPath := filepath.Join(t.TempDir(), "feature-x")
	if err := os.MkdirAll(filepath.Join(wtPath, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	porcelain := strings.TrimSpace(`
worktree `+wtPath+`
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
`) + "\n"

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res: runner.Result{
						ExitCode: 0,
					},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"merge-base", "--is-ancestor", "refs/heads/feature-x", "main"},
					res: runner.Result{
						ExitCode: 0,
					},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got []jsonWorktree
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	if got[0].PathExists == nil || got[0].DotGitExists == nil || got[0].Valid == nil || got[0].MergedIntoBase == nil {
		t.Fatalf("expected verify fields to be present: %#v", got[0])
	}
	if got[0].BaseRef != "main" {
		t.Fatalf("baseRef = %q, want main", got[0].BaseRef)
	}
	if *got[0].MergedIntoBase != true {
		t.Fatalf("mergedIntoBase = %v, want true", *got[0].MergedIntoBase)
	}
}
