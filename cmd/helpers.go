package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

// newSpinner creates a consistently configured spinner.
func newSpinner(w io.Writer, suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = w
	s.Suffix = fmt.Sprintf(" %s", suffix)
	return s
}

// registerItemTypeFlags adds the standard --issues, --prs, --discussions flags to a command.
func registerItemTypeFlags(cmd *cobra.Command, issues, prs, discussions *bool, verb string) {
	cmd.Flags().BoolVar(issues, "issues", false, fmt.Sprintf("%s only issues", verb))
	cmd.Flags().BoolVar(prs, "prs", false, fmt.Sprintf("%s only pull requests", verb))
	cmd.Flags().BoolVar(discussions, "discussions", false, fmt.Sprintf("%s only discussions", verb))
}
