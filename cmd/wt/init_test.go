package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestInit_Zsh(t *testing.T) {
	t.Parallel()

	cmd := newInitCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"zsh"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "wtr() { cd \"$(wt root)\" || return; }") {
		t.Fatalf("stdout = %q, want wtr function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "# Apply now: eval \"$(wt init zsh)\"") {
		t.Fatalf("stdout = %q, want zsh apply-now guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#   wt completion zsh > ~/.zsh/completions/_wt") {
		t.Fatalf("stdout = %q, want zsh completion install guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), "wtg() { cd \"$(wt path") {
		t.Fatalf("stdout = %q, want wtg function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "wcd() { cd \"$(wt path") {
		t.Fatalf("stdout = %q, want wcd function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "compdef _wtg wtg wcd") {
		t.Fatalf("stdout = %q, want completion bridge for wtg and wcd", stdout.String())
	}
}

func TestInit_Fish(t *testing.T) {
	t.Parallel()

	cmd := newInitCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"fish"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "function wtr") {
		t.Fatalf("stdout = %q, want fish wtr function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "# Apply now: wt init fish | source") {
		t.Fatalf("stdout = %q, want fish apply-now guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#   wt completion fish > ~/.config/fish/completions/wt.fish") {
		t.Fatalf("stdout = %q, want fish completion install guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), "function wtg") {
		t.Fatalf("stdout = %q, want fish function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "function wcd") {
		t.Fatalf("stdout = %q, want fish wcd function", stdout.String())
	}
}

func TestInit_Bash(t *testing.T) {
	t.Parallel()

	cmd := newInitCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"bash"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), `wtr() { cd "$(wt root)" || return; }`) {
		t.Fatalf("stdout = %q, want bash wtr function", stdout.String())
	}
	if !strings.Contains(stdout.String(), `# Apply now: eval "$(wt init bash)"`) {
		t.Fatalf("stdout = %q, want bash apply-now guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), "#   wt completion bash > ~/.bash_completion.d/wt") {
		t.Fatalf("stdout = %q, want bash completion install guide", stdout.String())
	}
	if !strings.Contains(stdout.String(), `wtg() { cd "$(wt path "$@")" || return; }`) {
		t.Fatalf("stdout = %q, want bash wtg function", stdout.String())
	}
	if !strings.Contains(stdout.String(), `wcd() { cd "$(wt path "$@")" || return; }`) {
		t.Fatalf("stdout = %q, want bash wcd function", stdout.String())
	}
}
