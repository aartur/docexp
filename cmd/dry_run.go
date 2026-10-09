package cmd

import (
	"fmt"

	"github.com/aartur/docexp/internal/api"
	"github.com/spf13/cobra"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "parse the input documents and print the input of an expander script (as JSONL)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			lines, err := api.CalcExpanderScriptInput(arg)
			if err != nil {
				return err
			}
			for _, line := range lines {
				fmt.Println(line)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dryRunCmd)
}
