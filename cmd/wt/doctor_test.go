package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/runner"
)

func newDoctorCmdWithDeps(t *testing.T, d *deps) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	cmd := newDoctorCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.WithValue(context.Background(), depsKey{}, d))
	return cmd, &stdout, &stderr
}

func mapGetenv(values map[string]string) func(string) string {
	return func(key string) string {
		if v, ok := values[key]; ok {
			return v
		}
		return ""
	}
}

func TestDoctor_JSON_InRepoWithOverrides(t *testing.T) {
	t.Parallel()

	const cwd = "/repo/.wt/feature-x"
	const repo = "/repo"

	cmd, stdout, stderr := newDoctorCmdWithDeps(t, &deps{
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
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"config", "--local", "--get", "wt.root"},
					res:     runner.Result{Stdout: []byte(".trees\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("https://github.com/es5h/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: cwd,
					name:    "/mock/bin/gh",
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
				{
					workDir: cwd,
					name:    "/mock/bin/glab",
					args:    []string{"auth", "status"},
					res:     runner.Result{ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
		Getenv: mapGetenv(map[string]string{
			"WT_ROOT":   "custom/.wt",
			"WT_GH_BIN": "/mock/bin/gh",
			"SHELL":     "/bin/zsh",
			"HOME":      "/home/user",
		}),
		LookPath: func(file string) (string, error) {
			if file == "glab" {
				return "/mock/bin/glab", nil
			}
			return "", errors.New("not found")
		},
		ReadFile: func(path string) ([]byte, error) {
			if path == "/home/user/.zshrc" {
				return []byte("eval \"$(wt init zsh)\"\n"), nil
			}
			return nil, errors.New("not found")
		},
		FileExists: func(path string) bool {
			return path == "/home/user/.zsh/completions/_wt"
		},
	})

	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got doctorReport
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}

	if findDoctorStatus(got, "worktree.root.env") != doctorStatusOK {
		t.Fatalf("worktree.root.env status = %q, want %q", findDoctorStatus(got, "worktree.root.env"), doctorStatusOK)
	}
	if findDoctorStatus(got, "worktree.root.config") != doctorStatusOK {
		t.Fatalf("worktree.root.config status = %q, want %q", findDoctorStatus(got, "worktree.root.config"), doctorStatusOK)
	}
	if findDoctorStatus(got, "shell.completion") != doctorStatusOK {
		t.Fatalf("shell.completion status = %q, want %q", findDoctorStatus(got, "shell.completion"), doctorStatusOK)
	}
}

func TestDoctor_Text_OutsideRepoStillRuns(t *testing.T) {
	t.Parallel()

	const cwd = "/tmp/not-a-repo"

	cmd, stdout, stderr := newDoctorCmdWithDeps(t, &deps{
		Runner: &fakeRunner{
			t: t,
			calls: []fakeCall{
				{
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--show-toplevel"},
					res:     runner.Result{Stderr: []byte("fatal: not a git repository"), ExitCode: 128},
					err:     errors.New("exit status 128"),
				},
			},
		},
		Cwd:    cwd,
		Getenv: mapGetenv(map[string]string{}),
		LookPath: func(file string) (string, error) {
			return "", errors.New("not found")
		},
	})

	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "[unavailable] git.context") {
		t.Fatalf("stdout = %q, want unavailable git.context", stdout.String())
	}
}

func TestDoctor_Text_MissingHostingCLIAndCompletion(t *testing.T) {
	t.Parallel()

	const cwd = "/repo"
	const repo = "/repo"

	cmd, stdout, stderr := newDoctorCmdWithDeps(t, &deps{
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
					workDir: cwd,
					name:    "git",
					args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
					res:     runner.Result{Stdout: []byte(repo + "/.git\n"), ExitCode: 0},
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"config", "--local", "--get", "wt.root"},
					res:     runner.Result{ExitCode: 1},
					err:     errors.New("exit status 1"),
				},
				{
					workDir: repo,
					name:    "git",
					args:    []string{"remote", "get-url", "origin"},
					res:     runner.Result{Stdout: []byte("https://gitlab.com/team/wt.git\n"), ExitCode: 0},
				},
			},
		},
		Cwd: cwd,
		Getenv: mapGetenv(map[string]string{
			"SHELL": "/bin/bash",
			"HOME":  "/home/user",
		}),
		LookPath: func(file string) (string, error) {
			return "", errors.New("not found")
		},
		ReadFile: func(path string) ([]byte, error) {
			if path == "/home/user/.bashrc" {
				return []byte("# no wt helpers yet\n"), nil
			}
			return nil, errors.New("not found")
		},
		FileExists: func(path string) bool { return false },
	})

	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "[unavailable] hosting.glab: glab is required for current hosting provider") {
		t.Fatalf("stdout = %q, want missing glab guidance", stdout.String())
	}
	if !strings.Contains(stdout.String(), "[warn] shell.completion: completion file is missing") {
		t.Fatalf("stdout = %q, want missing completion warning", stdout.String())
	}
}

func findDoctorStatus(report doctorReport, name string) doctorStatus {
	for _, check := range report.Checks {
		if check.Name == name {
			return check.Status
		}
	}
	return ""
}
