package hosting

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"wt/internal/runner"
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

func TestFindGitLabCLI_UsesExplicitEnv(t *testing.T) {
	tempDir := t.TempDir()
	glabPath := filepath.Join(tempDir, "glab")
	if err := writeExecutableStub(glabPath); err != nil {
		t.Fatalf("writeExecutableStub() error = %v", err)
	}

	t.Setenv("WT_GLAB_BIN", glabPath)
	t.Setenv("PATH", "")
	t.Setenv("GOPATH", "")
	t.Setenv("HOME", tempDir)

	got, ok := findGitLabCLI()
	if !ok {
		t.Fatalf("findGitLabCLI() ok = false, want true")
	}
	if got != glabPath {
		t.Fatalf("findGitLabCLI() = %q, want %q", got, glabPath)
	}
}

func TestFindGitHubCLI_ReturnsFalseWhenNotOnPathOrExplicit(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("WT_GH_BIN", "")
	t.Setenv("PATH", tempDir)

	_, ok := findGitHubCLI()
	if ok {
		t.Fatalf("findGitHubCLI() ok = true, want false")
	}
}

func TestFindGitLabCLI_ReturnsFalseWhenNotOnPathOrExplicit(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("WT_GLAB_BIN", "")
	t.Setenv("PATH", tempDir)

	_, ok := findGitLabCLI()
	if ok {
		t.Fatalf("findGitLabCLI() ok = true, want false")
	}
}

func TestVerifyMerged_UnsupportedProviderDegrades(t *testing.T) {
	got, err := VerifyMerged(context.Background(), runner.OSRunner{}, t.TempDir(), ProviderUnknown, "feature-x", "main")
	if err != nil {
		t.Fatalf("VerifyMerged() error = %v", err)
	}
	if got.Provider != ProviderUnknown {
		t.Fatalf("Provider = %q, want %q", got.Provider, ProviderUnknown)
	}
	if got.Kind != "unknown" {
		t.Fatalf("Kind = %q, want unknown", got.Kind)
	}
	if got.Merged != nil {
		t.Fatalf("Merged = %#v, want nil", got.Merged)
	}
	if got.Reason != "unsupported-provider" {
		t.Fatalf("Reason = %q, want unsupported-provider", got.Reason)
	}
}
