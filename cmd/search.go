package cmd

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-md/internal/executil"
	"github.com/jackchuka/gh-md/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchRepo        string
	searchIssues      bool
	searchPRs         bool
	searchDiscussions bool
	searchList        bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search local gh-md files with FZF",
	Long: `Search through locally-pulled GitHub issues, PRs, and discussions using FZF.

The search opens an interactive FZF interface with a preview pane. After selecting
an item, you can choose an action: open in editor, push changes, view in browser,
copy file path, or pull fresh from GitHub.

Examples:
  gh md search                       # Search all local files
  gh md search bug                   # Search with initial query "bug"
  gh md search --repo owner/repo     # Search only in a specific repo
  gh md search --issues              # Search only issues
  gh md search --prs                 # Search only pull requests
  gh md search --discussions         # Search only discussions
  gh md search --list                # Print matches without FZF`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchRepo, "repo", "", "Limit search to a specific repo (owner/repo)")
	searchCmd.Flags().BoolVar(&searchIssues, "issues", false, "Search only issues")
	searchCmd.Flags().BoolVar(&searchPRs, "prs", false, "Search only pull requests")
	searchCmd.Flags().BoolVar(&searchDiscussions, "discussions", false, "Search only discussions")
	searchCmd.Flags().BoolVar(&searchList, "list", false, "Print matches without interactive FZF")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Check for FZF unless using --list mode
	if !searchList {
		if err := search.CheckFZFInstalled(); err != nil {
			return err
		}
	}

	// Build filters
	filters := search.Filters{
		Repo:        searchRepo,
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

	// Get initial query from args
	var query string
	if len(args) > 0 {
		query = args[0]
	}

	// List mode - just print and exit
	if searchList {
		output := search.FormatItemsForList(items)
		if query != "" {
			// Simple substring filter
			var filtered []search.Item
			queryLower := strings.ToLower(query)
			for _, item := range items {
				searchable := strings.ToLower(fmt.Sprintf("%s/%s #%d %s %s %s",
					item.Owner, item.Repo, item.Number, item.Type, item.State, item.Title))
				if strings.Contains(searchable, queryLower) {
					filtered = append(filtered, item)
				}
			}
			items = filtered
			output = search.FormatItemsForList(items)
		}
		cmd.Print(output)
		return nil
	}

	// Interactive FZF selection
	selected, err := search.RunSelector(items, query)
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
