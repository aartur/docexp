package cmd

import (
	"github.com/aartur/docexp/internal/api"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the expansion process for the input Markdown files, modifying the input in-place",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			expander, err := cmd.Flags().GetString("expander")
			if err != nil {
				return err
			}

			timeout, err := cmd.Flags().GetInt("timeout")
			if err != nil {
				return err
			}

			err = api.ExpandIntoFile(arg, arg, expander, nil, timeout)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringP("expander", "e", "", "path to the expander script")
	runCmd.MarkFlagRequired("expander")

	runCmd.Flags().IntP("timeout", "t", 30, "timeout of the whole run, in seconds")

	rootCmd.AddCommand(runCmd)
}
