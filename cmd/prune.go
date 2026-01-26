package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-md/internal/prune"
	"github.com/spf13/cobra"
)

var pruneConfirm bool

var pruneCmd = &cobra.Command{
	Use:   "prune [owner/repo]",
	Short: "Delete local markdown files for closed/merged items",
	Long: `Delete local markdown files for merged PRs and closed issues/PRs.

By default, this command performs a dry-run and lists files that would be deleted.
Use --confirm to actually delete the files.

State is determined from local file frontmatter only (no network requests).
If local state is stale, run 'gh md pull --include-closed' first to refresh.

Prunable items:
  - Issues: state == "closed"
  - Pull Requests: state == "merged" or state == "closed"
  - Discussions: state == "closed"

Examples:
  gh md prune                    # Dry-run: list files that would be deleted
  gh md prune --confirm          # Actually delete files
  gh md prune owner/repo         # Dry-run for specific repo only
  gh md prune owner/repo --confirm`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrune,
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVar(&pruneConfirm, "confirm", false, "Actually delete files (default is dry-run)")
}

func runPrune(cmd *cobra.Command, args []string) error {
	var repoFilter string
	if len(args) > 0 {
		repoFilter = args[0]
	}

	files, err := prune.FindPrunableFiles(repoFilter)
	if err != nil {
		return fmt.Errorf("failed to find prunable files: %w", err)
	}

	if len(files) == 0 {
		cmd.Println("No files to prune.")
		return nil
	}

	if pruneConfirm {
		// Actually delete files
		deleted, err := prune.DeleteFiles(files)
		if err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}

		cmd.Printf("Deleted %d files:\n\n", deleted)
		for _, f := range files {
			cmd.Printf("  %s (%s)\n", f.RelativePath(), f.State)
		}
	} else {
		// Dry-run: just list files
		cmd.Printf("Would delete %d files:\n\n", len(files))
		for _, f := range files {
			cmd.Printf("  %s (%s)\n", f.RelativePath(), f.State)
		}
		cmd.Println("\nRun with --confirm to delete these files.")
	}

	return nil
}
