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
			case "zsh":
				// Output-only: user must opt-in by pasting into their rc file.
				// Includes an optional completion bridge so `wtg <TAB>` completes like `wt goto <TAB>`.
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration (paste into your ~/.zshrc)")
				fmt.Fprintln(cmd.OutOrStdout(), `wtg() { cd "$(wt goto "$@")" || return; }`)
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "# Optional: completion bridge (requires `wt` zsh completion to be installed)")
				fmt.Fprintln(cmd.OutOrStdout(), `if whence -w compdef >/dev/null 2>&1; then`)
				fmt.Fprintln(cmd.OutOrStdout(), `  if ! (( $+functions[_wt] )); then`)
				fmt.Fprintln(cmd.OutOrStdout(), `    whence -w _wt >/dev/null 2>&1 && autoload -Uz _wt`)
				fmt.Fprintln(cmd.OutOrStdout(), `  fi`)
				fmt.Fprintln(cmd.OutOrStdout(), `  if (( $+functions[_wt] )); then`)
				fmt.Fprintln(cmd.OutOrStdout(), `    _wtg() {`)
				fmt.Fprintln(cmd.OutOrStdout(), `      local -a wtg_words`)
				fmt.Fprintln(cmd.OutOrStdout(), `      wtg_words=("${words[@]}")`)
				fmt.Fprintln(cmd.OutOrStdout(), `      words=(wt goto "${wtg_words[@]:1}")`)
				fmt.Fprintln(cmd.OutOrStdout(), `      (( CURRENT++ ))`)
				fmt.Fprintln(cmd.OutOrStdout(), `      _wt`)
				fmt.Fprintln(cmd.OutOrStdout(), `    }`)
				fmt.Fprintln(cmd.OutOrStdout(), `    compdef _wtg wtg`)
				fmt.Fprintln(cmd.OutOrStdout(), `  fi`)
				fmt.Fprintln(cmd.OutOrStdout(), `fi`)
				return nil
			case "bash":
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration (paste into your ~/.bashrc)")
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
