package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestUpgrade_DryRun(t *testing.T) {
	t.Parallel()

	d := &deps{Cwd: "/repo"}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"upgrade", "--dry-run"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, d))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "go install github.com/es5h/wt/cmd/wt@latest") {
		t.Fatalf("stderr = %q, want latest install command", stderr.String())
	}
}

func TestUpgrade_InvalidVersionAtPrefix(t *testing.T) {
	t.Parallel()

	d := &deps{Cwd: "/repo"}

	root := newRootCmd()
	root.SetArgs([]string{"upgrade", "--version", "@latest"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, d))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 || !strings.Contains(err.Error(), "must not include '@'") || !strings.Contains(err.Error(), "@latest") {
		t.Fatalf("err = %#v, want usage error for invalid version", err)
	}
}

func TestUpgrade_UsesCurrentBinaryDir(t *testing.T) {
	t.Parallel()

	var gotWorkDir, gotInstallDir, gotPackageRef string
	d := &deps{
		Cwd: "/repo",
		InstallWithGo: func(_ context.Context, workDir string, installDir string, packageRef string) (runner.Result, error) {
			gotWorkDir = workDir
			gotInstallDir = installDir
			gotPackageRef = packageRef
			return runner.Result{ExitCode: 0}, nil
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"upgrade", "--version", "v0.10.2"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, d))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if gotWorkDir != "/repo" {
		t.Fatalf("workDir = %q, want /repo", gotWorkDir)
	}
	if gotPackageRef != "github.com/es5h/wt/cmd/wt@v0.10.2" {
		t.Fatalf("packageRef = %q, want github.com/es5h/wt/cmd/wt@v0.10.2", gotPackageRef)
	}
	if gotInstallDir == "" || filepath.IsAbs(gotInstallDir) == false {
		t.Fatalf("installDir = %q, want absolute path", gotInstallDir)
	}
	if !strings.Contains(stdout.String(), "upgraded: github.com/es5h/wt/cmd/wt@v0.10.2") {
		t.Fatalf("stdout = %q, want upgraded message", stdout.String())
	}
}

func TestUpgrade_InstallerError(t *testing.T) {
	t.Parallel()

	d := &deps{
		Cwd: "/repo",
		InstallWithGo: func(_ context.Context, workDir string, installDir string, packageRef string) (runner.Result, error) {
			return runner.Result{Stderr: []byte("download failed\n"), ExitCode: 1}, fmt.Errorf("exit status 1")
		},
	}

	root := newRootCmd()
	root.SetArgs([]string{"upgrade"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, d))

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "go install failed: download failed") {
		t.Fatalf("err = %q, want installer stderr message", err.Error())
	}
}
