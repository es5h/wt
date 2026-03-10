package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/es5h/wt/internal/runner"
)

func TestRoot_PathOnly(t *testing.T) {
	t.Parallel()

	const cwd = "/repo/.wt/feature-x"
	const repo = "/repo"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte(repo + "/.git\n"),
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"root"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.String() != "/repo\n" {
		t.Fatalf("stdout = %q, want path-only output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRoot_JSON(t *testing.T) {
	t.Parallel()

	const cwd = "/repo/.wt/feature-x"
	const repo = "/repo"

	r := &fakeRunner{
		t: t,
		calls: []fakeCall{
			{
				workDir: cwd,
				name:    "git",
				args:    []string{"rev-parse", "--path-format=absolute", "--git-common-dir"},
				res: runner.Result{
					Stdout:   []byte(repo + "/.git\n"),
					ExitCode: 0,
				},
			},
		},
	}

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"root", "--json"})
	root.SetContext(context.WithValue(context.Background(), depsKey{}, &deps{Runner: r, Cwd: cwd}))

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got struct {
		Root string `json:"root"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid json: %v\nstdout=%q", err, stdout.String())
	}
	if got.Root != repo {
		t.Fatalf("root = %q, want %q", got.Root, repo)
	}
}
