package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestWorktreeRemove_RemovesReadonlyCacheEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("readonly permission semantics differ on windows")
	}

	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init", "-b", "main")
	runGit(t, repoRoot, "config", "user.name", "wt-test")
	runGit(t, repoRoot, "config", "user.email", "wt-test@example.com")
	writeFile(t, filepath.Join(repoRoot, "README.md"), []byte("main\n"), 0o644)
	runGit(t, repoRoot, "add", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
	runGit(t, repoRoot, "branch", "feature/readonly-cache")

	linkedRoot := filepath.Join(t.TempDir(), "feature-readonly-cache")
	runGit(t, repoRoot, "worktree", "add", linkedRoot, "feature/readonly-cache")

	readonlyDir := filepath.Join(linkedRoot, ".cache", "go", "pkg", "mod", "cache", "download")
	if err := os.MkdirAll(readonlyDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	readonlyFile := filepath.Join(readonlyDir, "sumdb.txt")
	writeFile(t, readonlyFile, []byte("cached\n"), 0o444)
	if err := os.Chmod(readonlyDir, 0o555); err != nil {
		t.Fatalf("os.Chmod(%q) error = %v", readonlyDir, err)
	}

	if err := WorktreeRemove(context.Background(), runner.OSRunner{}, repoRoot, linkedRoot, true); err != nil {
		t.Fatalf("WorktreeRemove() error = %v", err)
	}

	if _, err := os.Stat(linkedRoot); !os.IsNotExist(err) {
		t.Fatalf("worktree path still exists, stat err = %v", err)
	}

	list := runGit(t, repoRoot, "worktree", "list", "--porcelain")
	if strings.Contains(list, linkedRoot) {
		t.Fatalf("worktree list still contains removed path: %s", list)
	}
}

func TestWorktreeRemove_PermissionDeniedIncludesTargetPath(t *testing.T) {
	repoRoot := t.TempDir()
	target := filepath.Join(repoRoot, "wt-target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	r := &scriptedRunner{
		t: t,
		calls: []runnerCall{
			{
				workDir: repoRoot,
				name:    "git",
				args:    []string{"worktree", "remove", "--force", target},
				res:     runner.Result{ExitCode: 1, Stderr: []byte("permission denied")},
				err:     errors.New("exit 1"),
			},
		},
	}

	err := WorktreeRemove(context.Background(), r, repoRoot, target, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "permission denied") {
		t.Fatalf("err = %q, want permission denied", msg)
	}
	if !strings.Contains(msg, "target="+target) {
		t.Fatalf("err = %q, want target path", msg)
	}
}

type runnerCall struct {
	workDir string
	name    string
	args    []string
	res     runner.Result
	err     error
}

type scriptedRunner struct {
	t     *testing.T
	calls []runnerCall
	i     int
}

func (s *scriptedRunner) Run(_ context.Context, workDir string, name string, args ...string) (runner.Result, error) {
	s.t.Helper()

	if s.i >= len(s.calls) {
		s.t.Fatalf("unexpected command: dir=%q name=%q args=%q", workDir, name, args)
	}
	want := s.calls[s.i]
	s.i++

	if workDir != want.workDir || name != want.name {
		s.t.Fatalf("command mismatch: got dir=%q name=%q want dir=%q name=%q", workDir, name, want.workDir, want.name)
	}
	if len(args) != len(want.args) {
		s.t.Fatalf("args len mismatch: got=%d want=%d, got=%q want=%q", len(args), len(want.args), args, want.args)
	}
	for i := range args {
		if args[i] != want.args[i] {
			s.t.Fatalf("args[%d] mismatch: got=%q want=%q (all got=%q want=%q)", i, args[i], want.args[i], args, want.args)
		}
	}
	return want.res, want.err
}

func runGit(t *testing.T, workDir string, args ...string) string {
	t.Helper()
	res, err := runner.OSRunner{}.Run(context.Background(), workDir, "git", args...)
	if err != nil {
		t.Fatalf("git %s: %s", strings.Join(args, " "), string(res.Stderr))
	}
	return strings.TrimSpace(string(res.Stdout))
}

func writeFile(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}
