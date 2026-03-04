package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"wt/internal/runner"
)

type removeRunnerCall struct {
	workDir string
	name    string
	args    []string
	res     runner.Result
	err     error
}

type removeRunner struct {
	t     *testing.T
	calls []removeRunnerCall
	idx   int
}

func (r *removeRunner) Run(_ context.Context, workDir string, name string, args ...string) (runner.Result, error) {
	r.t.Helper()
	if r.idx >= len(r.calls) {
		r.t.Fatalf("unexpected call %s %v", name, args)
	}
	call := r.calls[r.idx]
	r.idx++

	if call.workDir != workDir {
		r.t.Fatalf("workDir mismatch: got %q want %q", workDir, call.workDir)
	}
	if call.name != name {
		r.t.Fatalf("name mismatch: got %q want %q", name, call.name)
	}
	if len(call.args) != len(args) {
		r.t.Fatalf("args length mismatch: got %d want %d", len(args), len(call.args))
	}
	for i := range args {
		if args[i] != call.args[i] {
			r.t.Fatalf("args[%d] mismatch: got %q want %q", i, args[i], call.args[i])
		}
	}
	return call.res, call.err
}

func TestWorktreeRemove_RetryAfterPermissionDenied(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	lockedDir := filepath.Join(target, ".cache", "go", "pkg", "mod")
	if err := os.MkdirAll(lockedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lockedFile := filepath.Join(lockedDir, "module.txt")
	if err := os.WriteFile(lockedFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chmod(lockedDir, 0o555); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	if err := os.Chmod(lockedFile, 0o444); err != nil {
		t.Fatalf("chmod file: %v", err)
	}

	r := &removeRunner{
		t: t,
		calls: []removeRunnerCall{
			{
				workDir: "/repo",
				name:    "git",
				args:    []string{"worktree", "remove", "--force", target},
				res:     runner.Result{Stderr: []byte("Permission denied"), ExitCode: 128},
				err:     errors.New("exit status 128"),
			},
			{
				workDir: "/repo",
				name:    "git",
				args:    []string{"worktree", "remove", "--force", target},
				res:     runner.Result{ExitCode: 0},
			},
		},
	}

	if err := WorktreeRemove(context.Background(), r, "/repo", target, true); err != nil {
		t.Fatalf("WorktreeRemove() error = %v", err)
	}
	if r.idx != len(r.calls) {
		t.Fatalf("calls = %d, want %d", r.idx, len(r.calls))
	}

	dirInfo, err := os.Stat(lockedDir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if dirInfo.Mode().Perm()&0o200 == 0 {
		t.Fatalf("dir mode = %o, want user writable", dirInfo.Mode().Perm())
	}

	fileInfo, err := os.Stat(lockedFile)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if fileInfo.Mode().Perm()&0o200 == 0 {
		t.Fatalf("file mode = %o, want user writable", fileInfo.Mode().Perm())
	}
}

func TestWorktreeRemove_NoRetryForNonPermissionError(t *testing.T) {
	t.Parallel()

	r := &removeRunner{
		t: t,
		calls: []removeRunnerCall{
			{
				workDir: "/repo",
				name:    "git",
				args:    []string{"worktree", "remove", "--force", "/repo/.wt/feature"},
				res:     runner.Result{Stderr: []byte("fatal: bad revision"), ExitCode: 128},
				err:     errors.New("exit status 128"),
			},
		},
	}

	err := WorktreeRemove(context.Background(), r, "/repo", "/repo/.wt/feature", true)
	if err == nil {
		t.Fatalf("WorktreeRemove() error = nil, want error")
	}
	if r.idx != 1 {
		t.Fatalf("calls = %d, want 1", r.idx)
	}
}
