package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-md/internal/executil"
	"github.com/jackchuka/gh-md/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchIssues      bool
	searchPRs         bool
	searchDiscussions bool
	searchList        bool
)

var searchCmd = &cobra.Command{
	Use:   "search [owner/repo]",
	Short: "Search local gh-md files with FZF",
	Long: `Search through locally-pulled GitHub issues, PRs, and discussions using FZF.

The search opens an interactive FZF interface with a preview pane. After selecting
an item, you can choose an action: open in editor, push changes, view in browser,
copy file path, or pull fresh from GitHub.

Examples:
  gh md search                       # Search all local files
  gh md search owner/repo            # Search only in a specific repo
  gh md search --issues              # Search only issues
  gh md search --prs                 # Search only pull requests
  gh md search --discussions         # Search only discussions
  gh md search --list                # Print matches without FZF`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	registerItemTypeFlags(searchCmd, &searchIssues, &searchPRs, &searchDiscussions, "Search")
	searchCmd.Flags().BoolVar(&searchList, "list", false, "Print matches without interactive FZF")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Check for FZF unless using --list mode
	if !searchList {
		if err := search.CheckFZFInstalled(); err != nil {
			return err
		}
	}

	// Get repo from positional argument
	var repo string
	if len(args) > 0 {
		repo = args[0]
	}

	// Build filters
	filters := search.Filters{
		Repo:        repo,
		Issues:      searchIssues,
		PRs:         searchPRs,
		Discussions: searchDiscussions,
	}

	// Discover local files
	items, err := search.DiscoverLocalFiles(filters)
	if err != nil {
		return fmt.Errorf("failed to discover local files: %w", err)
	}

	if len(items) == 0 {
		cmd.Println("No local files found. Run 'gh md pull' to download some first.")
		return nil
	}

	// List mode - just print and exit
	if searchList {
		cmd.Print(search.FormatItemsForList(items))
		return nil
	}

	// Interactive FZF selection
	selected, err := search.RunSelector(items, "")
	if err != nil {
		return err
	}

	if selected == nil {
		// User cancelled
		return nil
	}

	// Show action menu
	action, err := search.RunActionMenu(selected)
	if err != nil {
		return err
	}

	// Execute the action
	return executeAction(cmd, selected, action)
}

func executeAction(cmd *cobra.Command, item *search.Item, action search.Action) error {
	switch action {
	case search.ActionOpenEditor:
		return executil.OpenInEditor(item.FilePath)

	case search.ActionPush:
		return runPush(cmd, []string{item.FilePath})

	case search.ActionViewBrowser:
		return executil.OpenInBrowser(item.URL)

	case search.ActionCopyPath:
		if err := executil.CopyToClipboard(item.FilePath); err != nil {
			return err
		}
		cmd.Println("Copied to clipboard")
		return nil

	case search.ActionPullFresh:
		return runPull(cmd, []string{item.FilePath})

	case search.ActionCancel:
		return nil

	default:
		return nil
	}
}
