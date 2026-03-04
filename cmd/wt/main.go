package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"wt/internal/buildinfo"
	"wt/internal/runner"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:]))
}

func run(ctx context.Context, args []string) int {
	rootCmd := newRootCmd()
	rootCmd.SetArgs(args)
	rootCmd.SetContext(ctx)

	err := rootCmd.Execute()
	if err == nil {
		return 0
	}

	var exitErr *exitError
	if errors.As(err, &exitErr) {
		if exitErr.Err != nil {
			fmt.Fprintln(os.Stderr, exitErr.Err)
		}
		return exitErr.Code
	}

	fmt.Fprintln(os.Stderr, err)
	return 1
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "wt",
		Short:         "git worktree helper (WIP)",
		Version:       buildinfo.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureDeps(cmd)
		},
	}
	rootCmd.SetVersionTemplate("wt {{.Version}}\n")

	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newPathCmd())
	rootCmd.AddCommand(newRepoRootCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newPruneCmd())
	rootCmd.AddCommand(newRemoveCmd())
	return rootCmd
}

type depsKey struct{}

type deps struct {
	Runner        runner.Runner
	Cwd           string
	IsInteractive func() bool
}

func ensureDeps(cmd *cobra.Command) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Value(depsKey{}).(*deps); ok {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	d := &deps{
		Runner:        runner.OSRunner{Env: os.Environ()},
		Cwd:           cwd,
		IsInteractive: stdinIsTTY,
	}

	cmd.SetContext(context.WithValue(ctx, depsKey{}, d))
	return nil
}

func getDeps(cmd *cobra.Command) (*deps, error) {
	ctx := cmd.Context()
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	d, ok := ctx.Value(depsKey{}).(*deps)
	if !ok || d == nil {
		return nil, fmt.Errorf("missing dependencies (internal error)")
	}
	return d, nil
}

type exitError struct {
	Code int
	Err  error
}

func (e *exitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit %d", e.Code)
	}
	return e.Err.Error()
}

func usageError(err error) error {
	return &exitError{Code: 2, Err: err}
}
