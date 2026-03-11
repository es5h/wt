package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/git"
	"github.com/es5h/wt/internal/hosting"
)

type doctorStatus string

const (
	doctorStatusOK          doctorStatus = "ok"
	doctorStatusWarn        doctorStatus = "warn"
	doctorStatusUnavailable doctorStatus = "unavailable"
)

type doctorCheck struct {
	Name    string       `json:"name"`
	Status  doctorStatus `json:"status"`
	Summary string       `json:"summary"`
	Details []string     `json:"details,omitempty"`
}

type doctorReport struct {
	Checks []doctorCheck `json:"checks"`
}

type shellSetup struct {
	RCPath          string
	CompletionPaths []string
	InitMarkers     []string
	InitHint        string
	CompletionHint  string
}

func newDoctorCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Diagnose wt environment and setup",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := getDeps(cmd)
			if err != nil {
				return err
			}

			report := runDoctor(cmd.Context(), d)
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}

			printDoctorText(cmd.OutOrStdout(), report)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}

func runDoctor(ctx context.Context, d *deps) doctorReport {
	getenv := d.Getenv
	if getenv == nil {
		getenv = func(key string) string { return "" }
	}
	lookPath := d.LookPath
	if lookPath == nil {
		lookPath = func(file string) (string, error) { return "", fmt.Errorf("lookPath unavailable") }
	}
	readFile := d.ReadFile
	if readFile == nil {
		readFile = func(path string) ([]byte, error) { return nil, fmt.Errorf("readFile unavailable") }
	}
	fileExists := d.FileExists
	if fileExists == nil {
		fileExists = func(path string) bool { return false }
	}

	report := doctorReport{}

	repoRoot, repoErr := git.RepoRoot(ctx, d.Runner, d.Cwd)
	if repoErr != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "git.context",
			Status:  doctorStatusUnavailable,
			Summary: "not in a Git repository",
			Details: []string{repoErr.Error()},
		})
	} else {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "git.context",
			Status:  doctorStatusOK,
			Summary: "Git repository detected",
			Details: []string{"repo root: " + repoRoot},
		})
	}

	primaryRoot := ""
	if repoErr != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "git.primary-root",
			Status:  doctorStatusUnavailable,
			Summary: "cannot resolve primary root outside Git context",
		})
	} else {
		root, err := git.PrimaryWorktreeRoot(ctx, d.Runner, d.Cwd)
		if err != nil {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "git.primary-root",
				Status:  doctorStatusUnavailable,
				Summary: "failed to resolve primary root",
				Details: []string{err.Error()},
			})
		} else {
			primaryRoot = root
			summary := "primary root resolved"
			if strings.TrimSpace(repoRoot) != strings.TrimSpace(primaryRoot) {
				summary = "linked worktree context detected"
			}
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "git.primary-root",
				Status:  doctorStatusOK,
				Summary: summary,
				Details: []string{"primary root: " + primaryRoot},
			})
		}
	}

	wtRootEnv := strings.TrimSpace(getenv("WT_ROOT"))
	if wtRootEnv == "" {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "worktree.root.env",
			Status:  doctorStatusOK,
			Summary: "WT_ROOT is not set",
		})
	} else {
		details := []string{"WT_ROOT: " + wtRootEnv}
		if primaryRoot != "" {
			details = append(details, "resolved: "+normalizeWorktreeRoot(primaryRoot, wtRootEnv))
		}
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "worktree.root.env",
			Status:  doctorStatusOK,
			Summary: "WT_ROOT override is active",
			Details: details,
		})
	}

	wtRootConfig := ""
	if repoErr != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "worktree.root.config",
			Status:  doctorStatusUnavailable,
			Summary: "cannot read wt.root outside Git context",
		})
	} else {
		cfg, err := git.ConfigGetLocal(ctx, d.Runner, repoRoot, "wt.root")
		if err != nil {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "worktree.root.config",
				Status:  doctorStatusUnavailable,
				Summary: "failed to read repo-local wt.root",
				Details: []string{err.Error()},
			})
		} else if strings.TrimSpace(cfg) == "" {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "worktree.root.config",
				Status:  doctorStatusOK,
				Summary: "wt.root is not set",
			})
		} else {
			wtRootConfig = cfg
			details := []string{"wt.root: " + cfg}
			if primaryRoot != "" {
				details = append(details, "resolved: "+normalizeWorktreeRoot(primaryRoot, cfg))
			}
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "worktree.root.config",
				Status:  doctorStatusOK,
				Summary: "repo-local wt.root is configured",
				Details: details,
			})
		}
	}

	if primaryRoot == "" {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "worktree.root.effective",
			Status:  doctorStatusUnavailable,
			Summary: "cannot compute effective worktree root",
		})
	} else {
		source := "default"
		value := ".wt"
		switch {
		case wtRootEnv != "":
			source = "WT_ROOT"
			value = wtRootEnv
		case wtRootConfig != "":
			source = "wt.root"
			value = wtRootConfig
		}
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "worktree.root.effective",
			Status:  doctorStatusOK,
			Summary: fmt.Sprintf("effective root source: %s", source),
			Details: []string{"effective root: " + normalizeWorktreeRoot(primaryRoot, value)},
		})
	}

	provider := hosting.ProviderUnknown
	if repoErr != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "hosting.provider",
			Status:  doctorStatusUnavailable,
			Summary: "cannot detect hosting provider outside Git context",
		})
	} else {
		remoteURL, err := git.RemoteURL(ctx, d.Runner, repoRoot, "origin")
		if err != nil {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "hosting.provider",
				Status:  doctorStatusUnavailable,
				Summary: "failed to resolve origin remote",
				Details: []string{err.Error()},
			})
		} else if strings.TrimSpace(remoteURL) == "" {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "hosting.provider",
				Status:  doctorStatusWarn,
				Summary: "origin remote is not configured",
			})
		} else {
			provider = hosting.DetectProvider(remoteURL)
			if provider == hosting.ProviderUnknown {
				report.Checks = append(report.Checks, doctorCheck{
					Name:    "hosting.provider",
					Status:  doctorStatusWarn,
					Summary: "origin provider is not supported for hosting verify",
					Details: []string{"origin: " + remoteURL},
				})
			} else {
				report.Checks = append(report.Checks, doctorCheck{
					Name:    "hosting.provider",
					Status:  doctorStatusOK,
					Summary: fmt.Sprintf("origin provider: %s", provider),
					Details: []string{"origin: " + remoteURL},
				})
			}
		}
	}

	report.Checks = append(report.Checks, doctorAuthCheck(ctx, d, lookPath, getenv, "gh", "WT_GH_BIN", provider == hosting.ProviderGitHub))
	report.Checks = append(report.Checks, doctorAuthCheck(ctx, d, lookPath, getenv, "glab", "WT_GLAB_BIN", provider == hosting.ProviderGitLab))

	shellPath := strings.TrimSpace(getenv("SHELL"))
	if shellPath == "" {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.detect",
			Status:  doctorStatusWarn,
			Summary: "SHELL is not set",
		})
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.init",
			Status:  doctorStatusUnavailable,
			Summary: "shell init check skipped: shell not detected",
		})
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.completion",
			Status:  doctorStatusUnavailable,
			Summary: "completion check skipped: shell not detected",
		})
		return report
	}

	shellName := filepath.Base(shellPath)
	setup, supported := doctorShellSetup(getenv("HOME"), shellName)
	if !supported {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.detect",
			Status:  doctorStatusWarn,
			Summary: fmt.Sprintf("unsupported shell: %s", shellName),
			Details: []string{"supported: zsh, bash, fish"},
		})
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.init",
			Status:  doctorStatusUnavailable,
			Summary: "shell init check skipped: unsupported shell",
		})
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.completion",
			Status:  doctorStatusUnavailable,
			Summary: "completion check skipped: unsupported shell",
		})
		return report
	}

	report.Checks = append(report.Checks, doctorCheck{
		Name:    "shell.detect",
		Status:  doctorStatusOK,
		Summary: fmt.Sprintf("shell detected: %s", shellName),
		Details: []string{"path: " + shellPath},
	})

	rcBytes, rcErr := readFile(setup.RCPath)
	if rcErr != nil {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.init",
			Status:  doctorStatusWarn,
			Summary: "shell rc file is not readable; cannot verify wt init marker",
			Details: []string{
				"rc: " + setup.RCPath,
				"error: " + strings.TrimSpace(rcErr.Error()),
				setup.InitHint,
			},
		})
	} else {
		rcText := string(rcBytes)
		if containsAny(rcText, setup.InitMarkers) {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "shell.init",
				Status:  doctorStatusOK,
				Summary: "wt init marker found in rc file",
				Details: []string{"rc: " + setup.RCPath},
			})
		} else {
			report.Checks = append(report.Checks, doctorCheck{
				Name:    "shell.init",
				Status:  doctorStatusWarn,
				Summary: "wt init marker not found in rc file",
				Details: []string{
					"rc: " + setup.RCPath,
					setup.InitHint,
				},
			})
		}
	}

	foundCompletionPath := ""
	for _, candidate := range setup.CompletionPaths {
		if fileExists(candidate) {
			foundCompletionPath = candidate
			break
		}
	}

	if foundCompletionPath != "" {
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.completion",
			Status:  doctorStatusOK,
			Summary: "completion file found",
			Details: []string{"path: " + foundCompletionPath},
		})
	} else {
		details := make([]string, 0, len(setup.CompletionPaths)+1)
		for _, candidate := range setup.CompletionPaths {
			details = append(details, "checked: "+candidate)
		}
		details = append(details, setup.CompletionHint)
		report.Checks = append(report.Checks, doctorCheck{
			Name:    "shell.completion",
			Status:  doctorStatusWarn,
			Summary: "completion file not found in expected locations",
			Details: details,
		})
	}

	return report
}

func doctorAuthCheck(ctx context.Context, d *deps, lookPath func(file string) (string, error), getenv func(key string) string, cli string, envKey string, expected bool) doctorCheck {
	name := "hosting." + cli + ".auth"
	explicit := strings.TrimSpace(getenv(envKey))
	resolved := ""
	source := "PATH"
	if explicit != "" {
		resolved = explicit
		source = envKey
	} else if path, err := lookPath(cli); err == nil {
		resolved = path
	}

	if strings.TrimSpace(resolved) == "" {
		summary := cli + " not found"
		if expected {
			summary = cli + " is required for current hosting provider"
		} else {
			summary = cli + " not found (optional for current hosting provider)"
		}
		return doctorCheck{
			Name:    name,
			Status:  doctorStatusUnavailable,
			Summary: summary,
			Details: []string{fmt.Sprintf("set %s or install %s", envKey, cli)},
		}
	}

	res, err := d.Runner.Run(ctx, d.Cwd, resolved, "auth", "status")
	if err != nil {
		summary := cli + " found, but auth status failed"
		if expected {
			summary = cli + " found, but authentication is required or unavailable"
		}
		details := []string{fmt.Sprintf("source: %s", source), fmt.Sprintf("path: %s", resolved)}
		if msg := strings.TrimSpace(string(res.Stderr)); msg != "" {
			details = append(details, msg)
		}
		return doctorCheck{Name: name, Status: doctorStatusWarn, Summary: summary, Details: details}
	}

	return doctorCheck{
		Name:    name,
		Status:  doctorStatusOK,
		Summary: cli + " available and authenticated",
		Details: []string{fmt.Sprintf("source: %s", source), fmt.Sprintf("path: %s", resolved)},
	}
}

func doctorShellSetup(home string, shell string) (shellSetup, bool) {
	home = strings.TrimSpace(home)
	if home == "" {
		home = "~"
	}

	switch shell {
	case "zsh":
		return shellSetup{
			RCPath:          filepath.Join(home, ".zshrc"),
			CompletionPaths: []string{filepath.Join(home, ".zsh", "completions", "_wt"), filepath.Join(home, ".zfunc", "_wt")},
			InitMarkers:     []string{"wt init zsh", "wtr()", "wtg()"},
			InitHint:        "hint: eval \"$(wt init zsh)\"",
			CompletionHint:  "hint: wt completion zsh > ~/.zsh/completions/_wt",
		}, true
	case "bash":
		return shellSetup{
			RCPath:          filepath.Join(home, ".bashrc"),
			CompletionPaths: []string{filepath.Join(home, ".bash_completion.d", "wt"), filepath.Join(home, ".local", "share", "bash-completion", "completions", "wt")},
			InitMarkers:     []string{"wt init bash", "wtr()", "wtg()"},
			InitHint:        "hint: eval \"$(wt init bash)\"",
			CompletionHint:  "hint: wt completion bash > ~/.bash_completion.d/wt",
		}, true
	case "fish":
		return shellSetup{
			RCPath:          filepath.Join(home, ".config", "fish", "config.fish"),
			CompletionPaths: []string{filepath.Join(home, ".config", "fish", "completions", "wt.fish")},
			InitMarkers:     []string{"wt init fish", "function wtr", "function wtg"},
			InitHint:        "hint: eval (wt init fish)",
			CompletionHint:  "hint: wt completion fish > ~/.config/fish/completions/wt.fish",
		}, true
	default:
		return shellSetup{}, false
	}
}

func containsAny(text string, markers []string) bool {
	for _, marker := range markers {
		if strings.TrimSpace(marker) == "" {
			continue
		}
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func printDoctorText(out io.Writer, report doctorReport) {
	for _, check := range report.Checks {
		fmt.Fprintf(out, "[%s] %s: %s\n", check.Status, check.Name, check.Summary)
		for _, detail := range check.Details {
			fmt.Fprintf(out, "  - %s\n", detail)
		}
	}
}
