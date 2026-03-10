package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/es5h/wt/internal/git"
)

func newRepoRootCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:           "root",
		Short:         "Print primary repository root path",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := getDeps(cmd)
			if err != nil {
				return err
			}

			root, err := git.PrimaryWorktreeRoot(cmd.Context(), d.Runner, d.Cwd)
			if err != nil {
				return err
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(struct {
					Root string `json:"root"`
				}{
					Root: root,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout(), root)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	return cmd
}
