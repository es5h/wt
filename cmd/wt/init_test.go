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
	if !strings.Contains(stdout.String(), "wtg() { cd \"$(wt path") {
		t.Fatalf("stdout = %q, want wtg function", stdout.String())
	}
	if !strings.Contains(stdout.String(), "compdef _wtg wtg") {
		t.Fatalf("stdout = %q, want completion bridge", stdout.String())
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
	if !strings.Contains(stdout.String(), "function wtg") {
		t.Fatalf("stdout = %q, want fish function", stdout.String())
	}
}
