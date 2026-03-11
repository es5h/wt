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
	if findDoctorStatus(got, "hosting.gh.auth") != doctorStatusOK {
		t.Fatalf("hosting.gh.auth status = %q, want %q", findDoctorStatus(got, "hosting.gh.auth"), doctorStatusOK)
	}
	if findDoctorStatus(got, "hosting.glab.auth") != doctorStatusOK {
		t.Fatalf("hosting.glab.auth status = %q, want %q", findDoctorStatus(got, "hosting.glab.auth"), doctorStatusOK)
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
	if !strings.Contains(stdout.String(), "[unavailable] shell.init: shell init check skipped: shell not detected") {
		t.Fatalf("stdout = %q, want unavailable shell.init when shell is missing", stdout.String())
	}
	if !strings.Contains(stdout.String(), "[unavailable] shell.completion: completion check skipped: shell not detected") {
		t.Fatalf("stdout = %q, want unavailable shell.completion when shell is missing", stdout.String())
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
	if !strings.Contains(stdout.String(), "[unavailable] hosting.glab.auth: glab is required for current hosting provider") {
		t.Fatalf("stdout = %q, want missing glab guidance", stdout.String())
	}
	if !strings.Contains(stdout.String(), "[warn] shell.completion: completion file not found in expected locations") {
		t.Fatalf("stdout = %q, want missing completion warning", stdout.String())
	}
}

func TestDoctor_Text_HostingAuthWarnWhenRequired(t *testing.T) {
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
					res:     runner.Result{Stdout: []byte("https://github.com/es5h/wt.git\n"), ExitCode: 0},
				},
				{
					workDir: cwd,
					name:    "/mock/bin/gh",
					args:    []string{"auth", "status"},
					res:     runner.Result{Stderr: []byte("gh auth status failed"), ExitCode: 1},
					err:     errors.New("exit status 1"),
				},
			},
		},
		Cwd: cwd,
		Getenv: mapGetenv(map[string]string{
			"WT_GH_BIN": "/mock/bin/gh",
			"SHELL":     "/bin/zsh",
			"HOME":      "/home/user",
		}),
		LookPath: func(file string) (string, error) {
			return "", errors.New("not found")
		},
		ReadFile: func(path string) ([]byte, error) {
			if path == "/home/user/.zshrc" {
				return []byte("eval \"$(wt init zsh)\"\n"), nil
			}
			return nil, errors.New("not found")
		},
		FileExists: func(path string) bool {
			return path == "/home/user/.zfunc/_wt"
		},
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "[warn] hosting.gh.auth: gh found, but authentication is required or unavailable") {
		t.Fatalf("stdout = %q, want required auth warning", stdout.String())
	}
}

func TestDoctor_TextJSONParity_StatusAndChecks(t *testing.T) {
	t.Parallel()

	const cwd = "/repo"
	const repo = "/repo"
	const shell = "/bin/bash"
	const home = "/home/user"

	newDeps := func() *deps {
		return &deps{
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
				"SHELL":       shell,
				"HOME":        home,
				"WT_GLAB_BIN": "/mock/bin/glab",
			}),
			LookPath: func(file string) (string, error) {
				return "", errors.New("not found")
			},
			ReadFile: func(path string) ([]byte, error) {
				if path == "/home/user/.bashrc" {
					return []byte("# no wt init marker\n"), nil
				}
				return nil, errors.New("not found")
			},
			FileExists: func(path string) bool {
				return false
			},
		}
	}

	textCmd, textStdout, textStderr := newDoctorCmdWithDeps(t, newDeps())
	if err := textCmd.Execute(); err != nil {
		t.Fatalf("text Execute() error = %v", err)
	}
	if textStderr.Len() != 0 {
		t.Fatalf("text stderr = %q, want empty", textStderr.String())
	}
	textMap := parseDoctorTextStatuses(t, textStdout.String())

	jsonCmd, jsonStdout, jsonStderr := newDoctorCmdWithDeps(t, newDeps())
	jsonCmd.SetArgs([]string{"--json"})
	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("json Execute() error = %v", err)
	}
	if jsonStderr.Len() != 0 {
		t.Fatalf("json stderr = %q, want empty", jsonStderr.String())
	}
	var report doctorReport
	if err := json.Unmarshal(jsonStdout.Bytes(), &report); err != nil {
		t.Fatalf("json stdout is invalid: %v", err)
	}
	jsonMap := doctorStatusMap(report)

	if len(textMap) != len(jsonMap) {
		t.Fatalf("check count mismatch: text=%d json=%d\ntext=%v\njson=%v", len(textMap), len(jsonMap), textMap, jsonMap)
	}
	for name, jsonStatus := range jsonMap {
		textStatus, ok := textMap[name]
		if !ok {
			t.Fatalf("text output missing check %q", name)
		}
		if textStatus != jsonStatus {
			t.Fatalf("status mismatch for %q: text=%q json=%q", name, textStatus, jsonStatus)
		}
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

func doctorStatusMap(report doctorReport) map[string]doctorStatus {
	statusByName := make(map[string]doctorStatus, len(report.Checks))
	for _, check := range report.Checks {
		statusByName[check.Name] = check.Status
	}
	return statusByName
}

func parseDoctorTextStatuses(t *testing.T, text string) map[string]doctorStatus {
	t.Helper()

	statusByName := map[string]doctorStatus{}
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "[") {
			continue
		}
		closeIdx := strings.Index(line, "]")
		if closeIdx <= 1 {
			t.Fatalf("invalid doctor text status line: %q", line)
		}
		statusToken := line[1:closeIdx]
		rest := strings.TrimSpace(line[closeIdx+1:])
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid doctor text check line: %q", line)
		}
		name := strings.TrimSpace(parts[0])
		statusByName[name] = doctorStatus(statusToken)
	}

	if len(statusByName) == 0 {
		t.Fatalf("no status lines parsed from text output: %q", text)
	}
	return statusByName
}
