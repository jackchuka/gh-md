package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-md/internal/output"
	"github.com/jackchuka/gh-md/internal/prune"
	"github.com/spf13/cobra"
)

var (
	pruneConfirm bool
	pruneFormat  string
)

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
  gh md prune owner/repo --confirm
  gh md prune --format=json      # Output as JSON for scripting`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrune,
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().BoolVar(&pruneConfirm, "confirm", false, "Actually delete files (default is dry-run)")
	pruneCmd.Flags().StringVar(&pruneFormat, "format", "text", "Output format (text, json, yaml)")
}

// pruneResultOutput is the structured output for prune results.
type pruneResultOutput struct {
	DryRun  bool              `json:"dry_run" yaml:"dry_run"`
	Deleted int               `json:"deleted" yaml:"deleted"`
	Files   []pruneFileOutput `json:"files" yaml:"files"`
	Total   int               `json:"total" yaml:"total"`
}

type pruneFileOutput struct {
	Path     string `json:"path" yaml:"path"`
	ItemType string `json:"item_type" yaml:"item_type"`
	Number   int    `json:"number" yaml:"number"`
	State    string `json:"state" yaml:"state"`
	Owner    string `json:"owner" yaml:"owner"`
	Repo     string `json:"repo" yaml:"repo"`
}

func runPrune(cmd *cobra.Command, args []string) error {
	p := output.NewPrinter(cmd).WithFormat(output.ParseFormat(pruneFormat))

	var repoFilter string
	if len(args) > 0 {
		repoFilter = args[0]
	}

	files, err := prune.FindPrunableFiles(repoFilter)
	if err != nil {
		return fmt.Errorf("failed to find prunable files: %w", err)
	}

	if len(files) == 0 {
		if p.IsStructured() {
			return p.Structured(pruneResultOutput{
				DryRun:  !pruneConfirm,
				Deleted: 0,
				Files:   []pruneFileOutput{},
				Total:   0,
			})
		}
		p.Print("No files to prune.")
		return nil
	}

	// Convert to output format
	outputFiles := make([]pruneFileOutput, len(files))
	for i, f := range files {
		outputFiles[i] = pruneFileOutput{
			Path:     f.RelativePath(),
			ItemType: f.ItemType.Display(),
			Number:   f.Number,
			State:    f.State,
			Owner:    f.Owner,
			Repo:     f.Repo,
		}
	}

	if pruneConfirm {
		// Actually delete files
		deleted, err := prune.DeleteFiles(files)
		if err != nil {
			return fmt.Errorf("failed to delete files: %w", err)
		}

		if p.IsStructured() {
			return p.Structured(pruneResultOutput{
				DryRun:  false,
				Deleted: deleted,
				Files:   outputFiles,
				Total:   len(files),
			})
		}

		p.Printf("Deleted %d files:\n", deleted)
		for _, f := range files {
			p.Printf("  %s (%s)\n", f.RelativePath(), f.State)
		}
	} else {
		// Dry-run: just list files
		if p.IsStructured() {
			return p.Structured(pruneResultOutput{
				DryRun:  true,
				Deleted: 0,
				Files:   outputFiles,
				Total:   len(files),
			})
		}

		p.Printf("Would delete %d files:\n", len(files))
		for _, f := range files {
			p.Printf("  %s (%s)\n", f.RelativePath(), f.State)
		}
		p.Print("\nRun with --confirm to delete these files.")
	}

	return nil
}
