package hosting

import (
	"os"
	"path/filepath"
	"testing"
)

func writeExecutableStub(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		return err
	}
	return nil
}

func TestFindGitHubCLI_UsesExplicitEnv(t *testing.T) {
	tempDir := t.TempDir()
	ghPath := filepath.Join(tempDir, "gh")
	if err := writeExecutableStub(ghPath); err != nil {
		t.Fatalf("writeExecutableStub() error = %v", err)
	}

	t.Setenv("WT_GH_BIN", ghPath)
	t.Setenv("PATH", "")
	t.Setenv("GOPATH", "")
	t.Setenv("HOME", tempDir)

	got, ok := findGitHubCLI()
	if !ok {
		t.Fatalf("findGitHubCLI() ok = false, want true")
	}
	if got != ghPath {
		t.Fatalf("findGitHubCLI() = %q, want %q", got, ghPath)
	}
}

func TestFindGitHubCLI_UsesGopathFallback(t *testing.T) {
	tempDir := t.TempDir()
	ghPath := filepath.Join(tempDir, "bin", "gh")
	if err := writeExecutableStub(ghPath); err != nil {
		t.Fatalf("writeExecutableStub() error = %v", err)
	}

	t.Setenv("WT_GH_BIN", "")
	t.Setenv("PATH", "")
	t.Setenv("GOPATH", tempDir)
	t.Setenv("HOME", tempDir)

	got, ok := findGitHubCLI()
	if !ok {
		t.Fatalf("findGitHubCLI() ok = false, want true")
	}
	if got != ghPath {
		t.Fatalf("findGitHubCLI() = %q, want %q", got, ghPath)
	}
}

func TestFindGitHubCLI_UsesHomeGoBinFallback(t *testing.T) {
	tempDir := t.TempDir()
	ghPath := filepath.Join(tempDir, "go", "bin", "gh")
	if err := writeExecutableStub(ghPath); err != nil {
		t.Fatalf("writeExecutableStub() error = %v", err)
	}

	t.Setenv("WT_GH_BIN", "")
	t.Setenv("PATH", "")
	t.Setenv("GOPATH", "")
	t.Setenv("HOME", tempDir)

	got, ok := findGitHubCLI()
	if !ok {
		t.Fatalf("findGitHubCLI() ok = false, want true")
	}
	if got != ghPath {
		t.Fatalf("findGitHubCLI() = %q, want %q", got, ghPath)
	}
}
