package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docexp",
	Short: "docexp is a tool that expands code blocks in Markdown documents. See https://github.com/aartur/docexp.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
