package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/buildinfo"
)

type upgradeOpts struct {
	Version string
	DryRun  bool
}

func newUpgradeCmd() *cobra.Command {
	var opts upgradeOpts

	cmd := &cobra.Command{
		Use:           "upgrade",
		Short:         "Upgrade wt to a released version",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := getDeps(cmd)
			if err != nil {
				return err
			}

			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("wt upgrade: failed to resolve current executable path: %w", err)
			}
			installDir := filepath.Dir(execPath)
			if filepath.Base(execPath) != "wt" {
				if wtPath, lookErr := exec.LookPath("wt"); lookErr == nil && strings.TrimSpace(wtPath) != "" {
					installDir = filepath.Dir(wtPath)
				}
			}

			version := strings.TrimSpace(opts.Version)
			if version == "" {
				version = "latest"
			}
			if strings.HasPrefix(version, "@") {
				return usageError(fmt.Errorf("wt upgrade: --version must not include '@': %s", version))
			}
			if version == "latest" {
				resolver := d.ResolveLatestVersion
				if resolver == nil {
					resolver = resolveLatestVersion
				}
				resolvedVersion, resolveErr := resolver(cmd.Context(), d.Cwd, buildinfo.ModulePath)
				if resolveErr != nil {
					return fmt.Errorf("wt upgrade: failed to resolve latest release version: %w", resolveErr)
				}
				version = resolvedVersion
			}

			packageRef := fmt.Sprintf("%s/cmd/wt@%s", buildinfo.ModulePath, version)
			if opts.DryRun {
				fmt.Fprintf(cmd.ErrOrStderr(), "dry-run: GOBIN=%s go install %s\n", installDir, packageRef)
				return nil
			}

			installer := d.InstallWithGo
			if installer == nil {
				installer = installWithGo
			}

			res, err := installer(cmd.Context(), d.Cwd, installDir, packageRef)
			if err != nil {
				msg := strings.TrimSpace(string(res.Stderr))
				if msg == "" {
					msg = err.Error()
				}
				return fmt.Errorf("wt upgrade: go install failed: %s", msg)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "upgraded: %s\n", packageRef)
			fmt.Fprintf(cmd.OutOrStdout(), "install-dir: %s\n", installDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "latest", "release version to install (e.g. v0.10.2 or latest)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "print install command without executing it")
	return cmd
}
