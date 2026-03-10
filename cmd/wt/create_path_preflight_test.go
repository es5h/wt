package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestCreatePreflight_FailsForFileAndNonEmptyDir(t *testing.T) {
	t.Parallel()

	targetRoot := t.TempDir()
	filePath := filepath.Join(targetRoot, "existing-file")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	nonEmptyDir := filepath.Join(targetRoot, "non-empty-dir")
	if err := os.Mkdir(nonEmptyDir, 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		targetPath string
		wantErr    string
		runnerSeq  []fakeCall
	}{
		{
			name:       "create rejects file",
			args:       []string{"create", "feature-x", "--path", filePath},
			targetPath: filePath,
			wantErr:    "wt create: target path is an existing file: " + filePath,
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakePrimaryRootCall("/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
			},
		},
		{
			name:       "create rejects non-empty dir",
			args:       []string{"create", "feature-x", "--path", nonEmptyDir},
			targetPath: nonEmptyDir,
			wantErr:    "wt create: target path is a non-empty directory: " + nonEmptyDir,
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakePrimaryRootCall("/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
			},
		},
		{
			name:       "path --create rejects file",
			args:       []string{"path", "feature-x", "--create", "--path", filePath},
			targetPath: filePath,
			wantErr:    "wt path: target path is an existing file: " + filePath,
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
				fakePrimaryRootCall("/repo"),
			},
		},
		{
			name:       "path --create rejects non-empty dir",
			args:       []string{"path", "feature-x", "--create", "--path", nonEmptyDir},
			targetPath: nonEmptyDir,
			wantErr:    "wt path: target path is a non-empty directory: " + nonEmptyDir,
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
				fakePrimaryRootCall("/repo"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)
			root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
				Runner: &fakeRunner{t: t, calls: tt.runnerSeq},
				Cwd:    "/cwd",
			}))

			err := root.Execute()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var exitErr *exitError
			if !errors.As(err, &exitErr) || exitErr.Code != 2 {
				t.Fatalf("err = %#v, want exitError code 2", err)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("err = %q, want %q", err.Error(), tt.wantErr)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestCreatePreflight_DryRunAlsoRejectsInvalidTarget(t *testing.T) {
	t.Parallel()

	targetRoot := t.TempDir()
	filePath := filepath.Join(targetRoot, "existing-file")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr string
		calls   []fakeCall
	}{
		{
			name:    "create --dry-run rejects file",
			args:    []string{"create", "feature-x", "--dry-run", "--path", filePath},
			wantErr: "wt create: target path is an existing file: " + filePath,
			calls: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakePrimaryRootCall("/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
			},
		},
		{
			name:    "path --create --dry-run rejects file",
			args:    []string{"path", "feature-x", "--create", "--dry-run", "--path", filePath},
			wantErr: "wt path: target path is an existing file: " + filePath,
			calls: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
				fakePrimaryRootCall("/repo"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)
			root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
				Runner: &fakeRunner{t: t, calls: tt.calls},
				Cwd:    "/cwd",
			}))

			err := root.Execute()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var exitErr *exitError
			if !errors.As(err, &exitErr) || exitErr.Code != 2 {
				t.Fatalf("err = %#v, want exitError code 2", err)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("err = %q, want %q", err.Error(), tt.wantErr)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestCreatePreflight_AllowsEmptyDirectory(t *testing.T) {
	t.Parallel()

	targetRoot := t.TempDir()
	emptyDir := filepath.Join(targetRoot, "empty-dir")
	if err := os.Mkdir(emptyDir, 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		wantPath  string
		runnerSeq []fakeCall
	}{
		{
			name:     "create with empty directory path",
			args:     []string{"create", "feature-x", "--path", emptyDir},
			wantPath: emptyDir + "\n",
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakePrimaryRootCall("/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
				fakeLocalBranchExistsCall("/repo", "feature-x", true),
				fakeAttachExistingBranchCall("/repo", emptyDir, "feature-x"),
			},
		},
		{
			name:     "path --create with empty directory path",
			args:     []string{"path", "feature-x", "--create", "--path", emptyDir},
			wantPath: emptyDir + "\n",
			runnerSeq: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
				fakePrimaryRootCall("/repo"),
				fakeLocalBranchExistsCall("/repo", "feature-x", true),
				fakeAttachExistingBranchCall("/repo", emptyDir, "feature-x"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)
			root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
				Runner: &fakeRunner{t: t, calls: tt.runnerSeq},
				Cwd:    "/cwd",
			}))

			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if stdout.String() != tt.wantPath {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.wantPath)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestCreatePreflight_RejectsSymlinkConservatively(t *testing.T) {
	t.Parallel()

	targetRoot := t.TempDir()
	targetDir := filepath.Join(targetRoot, "target-dir")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}
	symlinkPath := filepath.Join(targetRoot, "target-link")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("os.Symlink() not supported: %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"create", "feature-x", "--path", symlinkPath})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				fakeRepoRootCall("/cwd", "/repo"),
				fakePrimaryRootCall("/repo"),
				fakeWorktreeListCall("/repo", preflightBasePorcelain),
			},
		},
		Cwd: "/cwd",
	}))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("err = %#v, want exitError code 2", err)
	}
	want := "wt create: target path is a symbolic link (unsupported): " + symlinkPath
	if err.Error() != want {
		t.Fatalf("err = %q, want %q", err.Error(), want)
	}
}

const preflightBasePorcelain = `
worktree /repo
HEAD 0123456789abcdef0123456789abcdef01234567
branch refs/heads/main
`

func fakeRepoRootCall(cwd string, repo string) fakeCall {
	return fakeCall{
		workDir: cwd,
		name:    "git",
		args:    []string{"rev-parse", "--show-toplevel"},
		res: runner.Result{
			Stdout:   []byte(repo + "\n"),
			ExitCode: 0,
		},
	}
}

func fakePrimaryRootCall(repo string) fakeCall {
	return fakeCall{
		workDir: repo,
		name:    "git",
		args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
		res: runner.Result{
			Stdout:   []byte(filepath.Join(repo, ".git") + "\n"),
			ExitCode: 0,
		},
	}
}

func fakeWorktreeListCall(repo string, porcelain string) fakeCall {
	return fakeCall{
		workDir: repo,
		name:    "git",
		args:    []string{"worktree", "list", "--porcelain"},
		res: runner.Result{
			Stdout:   []byte(strings.TrimSpace(porcelain) + "\n"),
			ExitCode: 0,
		},
	}
}

func fakeLocalBranchExistsCall(repo string, branch string, exists bool) fakeCall {
	exitCode := 1
	var err error = errors.New("exit 1")
	if exists {
		exitCode = 0
		err = nil
	}
	return fakeCall{
		workDir: repo,
		name:    "git",
		args:    []string{"rev-parse", "--verify", "--quiet", "refs/heads/" + branch + "^{commit}"},
		res: runner.Result{
			ExitCode: exitCode,
		},
		err: err,
	}
}

func fakeAttachExistingBranchCall(repo string, targetPath string, branch string) fakeCall {
	return fakeCall{
		workDir: repo,
		name:    "git",
		args:    []string{"worktree", "add", targetPath, branch},
		res: runner.Result{
			ExitCode: 0,
		},
	}
}
