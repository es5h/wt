package main

import (
	"context"
	"reflect"
	"testing"

	"wt/internal/runner"
)

type fakeCall struct {
	workDir string
	name    string
	args    []string
	res     runner.Result
	err     error
}

type fakeRunner struct {
	t     *testing.T
	calls []fakeCall
	i     int
}

func (f *fakeRunner) Run(_ context.Context, workDir string, name string, args ...string) (runner.Result, error) {
	f.t.Helper()

	if f.i >= len(f.calls) {
		f.t.Fatalf("unexpected command: dir=%q name=%q args=%q", workDir, name, args)
	}
	want := f.calls[f.i]
	f.i++

	if workDir != want.workDir || name != want.name || !reflect.DeepEqual(args, want.args) {
		f.t.Fatalf("command mismatch:\n  got:  dir=%q name=%q args=%q\n  want: dir=%q name=%q args=%q",
			workDir, name, args,
			want.workDir, want.name, want.args,
		)
	}

	return want.res, want.err
}
