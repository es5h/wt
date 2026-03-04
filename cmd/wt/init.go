package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "init <shell>",
		Short:         "Print shell integration snippet",
		Args:          cobra.ExactArgs(1),
		ValidArgs:     []string{"zsh", "bash", "fish"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := strings.ToLower(strings.TrimSpace(args[0]))
			switch shell {
			case "zsh", "bash":
				// Output-only: user must opt-in by pasting into their rc file.
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration (paste into your rc file)")
				fmt.Fprintln(cmd.OutOrStdout(), `wtg() { cd "$(wt goto "$@")" || return; }`)
				return nil
			case "fish":
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration (paste into config.fish)")
				fmt.Fprintln(cmd.OutOrStdout(), "function wtg")
				fmt.Fprintln(cmd.OutOrStdout(), "  cd (wt goto $argv); or return")
				fmt.Fprintln(cmd.OutOrStdout(), "end")
				return nil
			default:
				return usageError(fmt.Errorf("wt init: unsupported shell: %s", shell))
			}
		},
	}
	return cmd
}
