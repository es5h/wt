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

func decodeJSONObjects(t *testing.T, data []byte) []map[string]any {
	t.Helper()

	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, string(data))
	}
	return got
}

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
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res: runner.Result{
						Stdout:   []byte(repo + "/.git\n"),
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
	if !got[0].Current || !got[0].Primary || got[0].RecommendedAction != "none" || got[0].SafeToRemove {
		t.Fatalf("unexpected derived signals: %#v", got[0])
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

func TestList_VerifyHostingRejectsPorcelain(t *testing.T) {
	t.Parallel()

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
		Runner: &fakeRunner{t: t},
		Cwd:    "/cwd",
	})
	cmd.SetArgs([]string{"--porcelain", "--verify-hosting"})

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
		t.Fatalf("stderr = %q, want empty", stderr.String())
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res: runner.Result{
						Stdout:   []byte(repo + "/.git\n"),
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
	if !strings.Contains(stdout.String(), "merged") {
		t.Fatalf("stdout = %q, want merged marker", stdout.String())
	}
	if !strings.Contains(stdout.String(), "safe-remove") || !strings.Contains(stdout.String(), "recommend:remove") {
		t.Fatalf("stdout = %q, want remove recommendation markers", stdout.String())
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res: runner.Result{
						Stdout:   []byte(repo + "/.git\n"),
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

	got := decodeJSONObjects(t, stdout.Bytes())
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	for _, key := range []string{"pathExists", "dotGitExists", "valid", "mergedIntoBase", "baseRef"} {
		if _, ok := got[0][key]; !ok {
			t.Fatalf("expected verify field %q to be present: %#v", key, got[0])
		}
	}
	if got[0]["baseRef"] != "main" {
		t.Fatalf("baseRef = %#v, want main", got[0]["baseRef"])
	}
	if got[0]["mergedIntoBase"] != true {
		t.Fatalf("mergedIntoBase = %#v, want true", got[0]["mergedIntoBase"])
	}
	if got[0]["recommendedAction"] != "remove" || got[0]["safeToRemove"] != true || got[0]["stale"] != false {
		t.Fatalf("unexpected derived fields: %#v", got[0])
	}
}

func TestList_VerifyHosting_GitHubMergedPR(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const ghBin = "/mock/bin/gh"
	t.Setenv("WT_GH_BIN", ghBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@github.com:es5h/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"merge-base", "--is-ancestor", "refs/heads/feature-x", "main"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"pr", "list", "--state", "merged", "--head", "feature-x", "--json", "number,title,url", "--limit", "1", "--base", "main"},
					res:     runner.Result{Stdout: []byte(`[{"number":1,"title":"merged feature","url":"https://github.com/es5h/wt/pull/1"}]`), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"merge-base", "--is-ancestor", "refs/heads/feature-x", "main"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"pr", "list", "--state", "merged", "--head", "feature-x", "--json", "number,title,url", "--limit", "1", "--base", "main"},
					res:     runner.Result{Stdout: []byte(`[{"number":1,"title":"merged feature","url":"https://github.com/es5h/wt/pull/1"}]`), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--verify", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "merged-hosting:github") {
		t.Fatalf("stdout = %q, want merged-hosting:github marker", stdout.String())
	}
}

func TestList_VerifyHosting_JSONIncludesChangeMetadataOnly(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const ghBin = "/mock/bin/gh"
	t.Setenv("WT_GH_BIN", ghBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@github.com:es5h/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"pr", "list", "--state", "merged", "--head", "feature-x", "--json", "number,title,url", "--limit", "1", "--base", "main"},
					res:     runner.Result{Stdout: []byte(`[{"number":42,"title":"feature x","url":"https://github.com/es5h/wt/pull/42"}]`), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	if _, ok := got[0]["mergedIntoBase"]; ok {
		t.Fatalf("mergedIntoBase should be absent when only --verify-hosting is used: %#v", got[0])
	}
	if got[0]["mergedViaHosting"] != true {
		t.Fatalf("mergedViaHosting = %#v, want true", got[0]["mergedViaHosting"])
	}
	if got[0]["hostingChangeNumber"] != float64(42) {
		t.Fatalf("hostingChangeNumber = %#v, want 42", got[0]["hostingChangeNumber"])
	}
	if got[0]["hostingChangeTitle"] != "feature x" {
		t.Fatalf("hostingChangeTitle = %#v, want feature x", got[0]["hostingChangeTitle"])
	}
	if got[0]["hostingChangeUrl"] != "https://github.com/es5h/wt/pull/42" {
		t.Fatalf("hostingChangeUrl = %#v, want PR URL", got[0]["hostingChangeUrl"])
	}
	if got[0]["recommendedAction"] != "remove" || got[0]["safeToRemove"] != true || got[0]["stale"] != false {
		t.Fatalf("unexpected derived fields: %#v", got[0])
	}
}

func TestList_VerifyHosting_JSONUnavailable(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const ghBin = "/mock/bin/gh"
	t.Setenv("WT_GH_BIN", ghBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@github.com:es5h/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    ghBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	if got[0]["hostingProvider"] != "github" {
		t.Fatalf("hostingProvider = %#v, want github", got[0]["hostingProvider"])
	}
	if got[0]["hostingKind"] != "pr" {
		t.Fatalf("hostingKind = %#v, want pr", got[0]["hostingKind"])
	}
	if got[0]["mergedViaHosting"] != nil {
		t.Fatalf("mergedViaHosting = %#v, want nil", got[0]["mergedViaHosting"])
	}
	if got[0]["hostingReason"] != "gh-auth-unavailable" {
		t.Fatalf("hostingReason = %#v, want gh-auth-unavailable", got[0]["hostingReason"])
	}
	if _, ok := got[0]["mergedIntoBase"]; ok {
		t.Fatalf("mergedIntoBase should be absent when only --verify-hosting is used: %#v", got[0])
	}
}

func TestList_VerifyHosting_TextShowsNoteWhenGHUnavailable(t *testing.T) {
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

	t.Setenv("WT_GH_BIN", "")
	t.Setenv("PATH", "")

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
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
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@github.com:es5h/wt.git\n"), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "note: hosting verify skipped") {
		t.Fatalf("stderr = %q, want note", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatalf("stdout = empty, want list output")
	}
}

func TestList_VerifyHosting_JSONGitLabMergedMR(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const glabBin = "/mock/bin/glab"
	t.Setenv("WT_GLAB_BIN", glabBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@gitlab.com:team/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"api", "projects/:fullpath/merge_requests?state=merged&source_branch=feature-x&per_page=1&order_by=updated_at&sort=desc&target_branch=main"},
					res:     runner.Result{Stdout: []byte(`[{"iid":17,"title":"feature x","web_url":"https://gitlab.com/team/wt/-/merge_requests/17","merged_at":"2026-03-05T00:00:00Z"}]`), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if got[0]["hostingProvider"] != "gitlab" {
		t.Fatalf("hostingProvider = %#v, want gitlab", got[0]["hostingProvider"])
	}
	if got[0]["hostingKind"] != "mr" {
		t.Fatalf("hostingKind = %#v, want mr", got[0]["hostingKind"])
	}
	if got[0]["mergedViaHosting"] != true {
		t.Fatalf("mergedViaHosting = %#v, want true", got[0]["mergedViaHosting"])
	}
	if got[0]["hostingChangeNumber"] != float64(17) {
		t.Fatalf("hostingChangeNumber = %#v, want 17", got[0]["hostingChangeNumber"])
	}
	if got[0]["hostingChangeTitle"] != "feature x" {
		t.Fatalf("hostingChangeTitle = %#v, want feature x", got[0]["hostingChangeTitle"])
	}
	if got[0]["hostingChangeUrl"] != "https://gitlab.com/team/wt/-/merge_requests/17" {
		t.Fatalf("hostingChangeUrl = %#v, want MR URL", got[0]["hostingChangeUrl"])
	}
}

func TestList_VerifyHosting_JSONGitLabUnavailable(t *testing.T) {
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

	t.Setenv("WT_GLAB_BIN", "")
	t.Setenv("PATH", "")

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
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
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@gitlab.com:team/wt.git\n"), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if got[0]["hostingProvider"] != "gitlab" {
		t.Fatalf("hostingProvider = %#v, want gitlab", got[0]["hostingProvider"])
	}
	if got[0]["hostingKind"] != "mr" {
		t.Fatalf("hostingKind = %#v, want mr", got[0]["hostingKind"])
	}
	if got[0]["mergedViaHosting"] != nil {
		t.Fatalf("mergedViaHosting = %#v, want nil", got[0]["mergedViaHosting"])
	}
	if got[0]["hostingReason"] != "glab-auth-unavailable" {
		t.Fatalf("hostingReason = %#v, want glab-auth-unavailable", got[0]["hostingReason"])
	}
}

func TestList_VerifyHosting_JSONGitLabUnauthenticated(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const glabBin = "/mock/bin/glab"
	t.Setenv("WT_GLAB_BIN", glabBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@gitlab.com:team/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if got[0]["hostingReason"] != "glab-auth-unavailable" {
		t.Fatalf("hostingReason = %#v, want glab-auth-unavailable", got[0]["hostingReason"])
	}
}

func TestList_VerifyHosting_JSONGitLabQueryFailureDegrades(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const glabBin = "/mock/bin/glab"
	t.Setenv("WT_GLAB_BIN", glabBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@gitlab.com:team/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"api", "projects/:fullpath/merge_requests?state=merged&source_branch=feature-x&per_page=1&order_by=updated_at&sort=desc&target_branch=main"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--json", "--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	got := decodeJSONObjects(t, stdout.Bytes())
	if got[0]["mergedViaHosting"] != nil {
		t.Fatalf("mergedViaHosting = %#v, want nil", got[0]["mergedViaHosting"])
	}
	if got[0]["hostingReason"] != "glab-mr-query-failed" {
		t.Fatalf("hostingReason = %#v, want glab-mr-query-failed", got[0]["hostingReason"])
	}
}

func TestList_VerifyHosting_TextShowsNoteWhenGitLabQueryFails(t *testing.T) {
	const cwd = "/cwd"
	const repo = "/repo"
	const glabBin = "/mock/bin/glab"
	t.Setenv("WT_GLAB_BIN", glabBin)

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
					res:     runner.Result{Stdout: []byte(repo + "\n"), ExitCode: 0},
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
					args:    []string{"rev-parse", "--verify", "--quiet", "main^{commit}"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("git@gitlab.com:team/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"api", "projects/:fullpath/merge_requests?state=merged&source_branch=feature-x&per_page=1&order_by=updated_at&sort=desc&target_branch=main"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: repo,
					name:    glabBin,
					args:    []string{"api", "projects/:fullpath/merge_requests?state=merged&source_branch=feature-x&per_page=1&order_by=updated_at&sort=desc&target_branch=main"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit 1"),
				},
			},
		},
		Cwd: cwd,
	})

	cmd.SetArgs([]string{"--verify-hosting", "--base", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "note: hosting verify skipped (glab MR query failed)") {
		t.Fatalf("stderr = %q, want glab MR query failure note", stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatalf("stdout = empty, want list output")
	}
}

func TestList_Verify_JSONDetachedUsesNullMergedIntoBase(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	wtPath := filepath.Join(t.TempDir(), "detached")
	if err := os.MkdirAll(filepath.Join(wtPath, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	porcelain := strings.TrimSpace(`
worktree `+wtPath+`
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
detached
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
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res: runner.Result{
						Stdout:   []byte(repo + "/.git\n"),
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

	got := decodeJSONObjects(t, stdout.Bytes())
	if len(got) != 1 {
		t.Fatalf("len(json) = %d, want 1", len(got))
	}
	for _, key := range []string{"pathExists", "dotGitExists", "valid", "mergedIntoBase", "baseRef"} {
		if _, ok := got[0][key]; !ok {
			t.Fatalf("expected verify field %q to be present: %#v", key, got[0])
		}
	}
	if got[0]["mergedIntoBase"] != nil {
		t.Fatalf("mergedIntoBase = %#v, want null", got[0]["mergedIntoBase"])
	}
	if got[0]["baseRef"] != "main" {
		t.Fatalf("baseRef = %#v, want main", got[0]["baseRef"])
	}
	if got[0]["detached"] != true {
		t.Fatalf("detached = %#v, want true", got[0]["detached"])
	}
	if got[0]["recommendedAction"] != "none" || got[0]["safeToRemove"] != false {
		t.Fatalf("unexpected derived fields: %#v", got[0])
	}
}

func TestList_JSONPrunableSignals(t *testing.T) {
	t.Parallel()

	const cwd = "/cwd"
	const repo = "/repo"

	porcelain := strings.TrimSpace(`
worktree /repo/.wt/feature-x
HEAD abcdefabcdefabcdefabcdefabcdefabcdefabcd
branch refs/heads/feature-x
prunable gitdir file points to non-existent location
`) + "\n"

	cmd, stdout, stderr := newListCmdWithDeps(t, &deps{
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
				{
					workDir: repo,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
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

	got := decodeJSONObjects(t, stdout.Bytes())
	if got[0]["stale"] != true || got[0]["recommendedAction"] != "prune" || got[0]["safeToRemove"] != false {
		t.Fatalf("unexpected derived fields: %#v", got[0])
	}
}
