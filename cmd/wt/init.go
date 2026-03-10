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
				// Output-only: user must opt-in by evaluating or pasting this snippet.
				// Includes an optional completion bridge so `wtg <TAB>` and `wcd <TAB>` complete like `wt path <TAB>`.
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration for zsh (output-only; no rc changes are made)")
				fmt.Fprintln(cmd.OutOrStdout(), `# Apply now: eval "$(wt init zsh)"`)
				fmt.Fprintln(cmd.OutOrStdout(), `# Persist helpers: wt init zsh >> ~/.zshrc`)
				fmt.Fprintln(cmd.OutOrStdout(), "# Optional completion install:")
				fmt.Fprintln(cmd.OutOrStdout(), "#   mkdir -p ~/.zsh/completions")
				fmt.Fprintln(cmd.OutOrStdout(), "#   wt completion zsh > ~/.zsh/completions/_wt")
				fmt.Fprintln(cmd.OutOrStdout(), "#   fpath=(~/.zsh/completions $fpath)")
				fmt.Fprintln(cmd.OutOrStdout(), "#   autoload -Uz compinit && compinit")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), `wtr() { cd "$(wt root)" || return; }`)
				fmt.Fprintln(cmd.OutOrStdout(), `wtg() { cd "$(wt path "$@")" || return; }`)
				fmt.Fprintln(cmd.OutOrStdout(), `wcd() { cd "$(wt path "$@")" || return; }`)
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
				fmt.Fprintln(cmd.OutOrStdout(), `      words=(wt path "${wtg_words[@]:1}")`)
				fmt.Fprintln(cmd.OutOrStdout(), `      (( CURRENT++ ))`)
				fmt.Fprintln(cmd.OutOrStdout(), `      _wt`)
				fmt.Fprintln(cmd.OutOrStdout(), `    }`)
				fmt.Fprintln(cmd.OutOrStdout(), `    compdef _wtg wtg wcd`)
				fmt.Fprintln(cmd.OutOrStdout(), `  fi`)
				fmt.Fprintln(cmd.OutOrStdout(), `fi`)
				return nil
			case "bash":
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration for bash (output-only; no rc changes are made)")
				fmt.Fprintln(cmd.OutOrStdout(), `# Apply now: eval "$(wt init bash)"`)
				fmt.Fprintln(cmd.OutOrStdout(), `# Persist helpers: wt init bash >> ~/.bashrc`)
				fmt.Fprintln(cmd.OutOrStdout(), "# Optional completion install:")
				fmt.Fprintln(cmd.OutOrStdout(), "#   mkdir -p ~/.bash_completion.d")
				fmt.Fprintln(cmd.OutOrStdout(), "#   wt completion bash > ~/.bash_completion.d/wt")
				fmt.Fprintln(cmd.OutOrStdout(), "#   source ~/.bash_completion.d/wt")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), `wtr() { cd "$(wt root)" || return; }`)
				fmt.Fprintln(cmd.OutOrStdout(), `wtg() { cd "$(wt path "$@")" || return; }`)
				fmt.Fprintln(cmd.OutOrStdout(), `wcd() { cd "$(wt path "$@")" || return; }`)
				return nil
			case "fish":
				fmt.Fprintln(cmd.OutOrStdout(), "# wt shell integration for fish (output-only; no config changes are made)")
				fmt.Fprintln(cmd.OutOrStdout(), "# Apply now: wt init fish | source")
				fmt.Fprintln(cmd.OutOrStdout(), "# Persist helpers: wt init fish >> ~/.config/fish/config.fish")
				fmt.Fprintln(cmd.OutOrStdout(), "# Optional completion install:")
				fmt.Fprintln(cmd.OutOrStdout(), "#   mkdir -p ~/.config/fish/completions")
				fmt.Fprintln(cmd.OutOrStdout(), "#   wt completion fish > ~/.config/fish/completions/wt.fish")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "function wtr")
				fmt.Fprintln(cmd.OutOrStdout(), "  cd (wt root); or return")
				fmt.Fprintln(cmd.OutOrStdout(), "end")
				fmt.Fprintln(cmd.OutOrStdout(), "function wtg")
				fmt.Fprintln(cmd.OutOrStdout(), "  cd (wt path $argv); or return")
				fmt.Fprintln(cmd.OutOrStdout(), "end")
				fmt.Fprintln(cmd.OutOrStdout(), "function wcd")
				fmt.Fprintln(cmd.OutOrStdout(), "  cd (wt path $argv); or return")
				fmt.Fprintln(cmd.OutOrStdout(), "end")
				return nil
			default:
				return usageError(fmt.Errorf("wt init: unsupported shell: %s", shell))
			}
		},
	}
	return cmd
}
