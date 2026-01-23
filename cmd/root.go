package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "md",
	Short: "Convert GitHub data to local markdown files",
	Long: `gh-md is a GitHub CLI extension that converts GitHub data
(Issues, Discussions, Pull Requests) to local markdown files,
making them easy to browse and feed to AI agents.

Data is stored in ~/.gh-md/ (configurable via GH_MD_ROOT env var).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
