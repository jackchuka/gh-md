package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackchuka/gh-md/internal/executil"
	"github.com/jackchuka/gh-md/internal/meta"
	"github.com/jackchuka/gh-md/internal/output"
	"github.com/jackchuka/gh-md/internal/parser"
	"github.com/jackchuka/gh-md/internal/search"
	"github.com/spf13/cobra"
)

var (
	rootIssues      bool
	rootPRs         bool
	rootDiscussions bool
	rootFilter      string
	rootNew         bool
	rootAssigned    bool
	rootList        bool
	rootFormat      string
)

var rootCmd = &cobra.Command{
	Use:   "md [owner/repo]",
	Short: "Search and browse local GitHub markdown files",
	Long: `gh-md is a GitHub CLI extension that converts GitHub data
(Issues, Discussions, Pull Requests) to local markdown files,
making them easy to browse and feed to AI agents.

Data is stored in ~/.gh-md/ (configurable via GH_MD_ROOT env var).

When run without a subcommand, opens an interactive FZF selector to browse
local files. Use flags to filter:

  gh md                              # All local files
  gh md --new                        # Items updated since last pull
  gh md --assigned                   # Items assigned to you
  gh md --issues                     # Only issues
  gh md --filter 'state == "open"'   # CEL filter expression
  gh md owner/repo                   # Filter to specific repo
  gh md --list                       # Print matches without FZF
  gh md --list --format=json         # Output as JSON for scripting

CEL filter variables:
  user, now, item_type, state, title, body, author,
  assigned, reviewers, labels, created, updated, owner, repo, number`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRoot,
}

func init() {
	registerItemTypeFlags(rootCmd, &rootIssues, &rootPRs, &rootDiscussions, "Show")
	rootCmd.Flags().StringVar(&rootFilter, "filter", "", "CEL filter expression")
	rootCmd.Flags().BoolVar(&rootNew, "new", false, "Show items updated since last pull")
	rootCmd.Flags().BoolVar(&rootAssigned, "assigned", false, "Show items assigned to you")
	rootCmd.Flags().BoolVar(&rootList, "list", false, "Print matches without interactive FZF")
	rootCmd.Flags().StringVar(&rootFormat, "format", "text", "Output format: text, json, yaml (only with --list)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Check for FZF unless using --list mode
	if !rootList {
		if err := search.CheckFZFInstalled(); err != nil {
			return err
		}
	}

	// Get repo from positional argument
	var repo string
	if len(args) > 0 {
		repo = args[0]
	}

	var items []search.Item
	var err error

	// Determine which discovery method to use
	useCEL := rootFilter != "" || rootAssigned
	useNew := rootNew

	if useCEL {
		items, err = discoverWithCEL(cmd, repo)
	} else if useNew {
		items, err = discoverNewItems(cmd, repo)
	} else {
		items, err = discoverWithFilters(repo)
	}

	if err != nil {
		return err
	}

	p := output.NewPrinter(cmd).WithFormat(output.ParseFormat(rootFormat))

	if len(items) == 0 {
		p.Print("No items found. Run 'gh md pull' to download some first.")
		return nil
	}

	// List mode - just print and exit
	if rootList {
		return outputItems(p, items)
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

func discoverWithCEL(cmd *cobra.Command, repo string) ([]search.Item, error) {
	// Build CEL filter expression
	filterExpr := rootFilter
	if rootAssigned {
		if filterExpr != "" {
			filterExpr = fmt.Sprintf("(%s) && user in assigned", filterExpr)
		} else {
			filterExpr = "user in assigned"
		}
	}

	// Get current GitHub username
	s := newSpinner(cmd.ErrOrStderr(), "Getting GitHub username...")
	s.Start()
	username, err := search.GetCurrentUser()
	s.Stop()
	if err != nil {
		return nil, err
	}

	// Compile CEL filter
	s.Suffix = " Compiling filter..."
	s.Start()
	prg, err := search.CompileCELFilter(filterExpr)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("invalid filter expression: %w", err)
	}

	// Discover matching items
	s.Suffix = " Scanning local files..."
	s.Start()
	items, err := search.DiscoverItems(prg, username, repo)
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	return items, nil
}

func discoverNewItems(cmd *cobra.Command, repo string) ([]search.Item, error) {
	// Collect sync timestamps per repo
	syncTimes := make(map[string]*meta.SyncTimestamps)

	s := newSpinner(cmd.ErrOrStderr(), "Scanning local files...")
	s.Start()

	var items []search.Item

	err := parser.WalkParsedFiles(parser.WalkFilters{Repo: repo}, func(parsed *parser.ParsedFile) error {
		itemType := "unknown"
		if label, ok := parsed.ItemType.ListLabel(); ok {
			itemType = label
		}

		// Apply type filters
		if rootIssues || rootPRs || rootDiscussions {
			switch itemType {
			case "issue":
				if !rootIssues {
					return nil
				}
			case "pr":
				if !rootPRs {
					return nil
				}
			case "discussion":
				if !rootDiscussions {
					return nil
				}
			}
		}

		// Get sync timestamp for this repo
		repoKey := parsed.Owner + "/" + parsed.Repo
		if _, ok := syncTimes[repoKey]; !ok {
			m, err := meta.Load(parsed.Owner, parsed.Repo)
			if err != nil {
				syncTimes[repoKey] = nil
			} else {
				syncTimes[repoKey] = m.Sync
			}
		}

		// Check if item is newer than previous sync (for --new flag)
		// This shows items that were updated between the previous pull and the current one
		sync := syncTimes[repoKey]
		if sync != nil {
			var prevSync *time.Time
			switch itemType {
			case "issue":
				prevSync = sync.PrevIssues
			case "pr":
				prevSync = sync.PrevPulls
			case "discussion":
				prevSync = sync.PrevDiscussions
			}
			// If prevSync exists, filter to items updated after it
			// If prevSync is nil (first pull), show all items
			if prevSync != nil && !parsed.Updated.After(*prevSync) {
				return nil
			}
		}

		url := ""
		if seg, ok := parsed.ItemType.URLSegment(); ok {
			url = fmt.Sprintf("https://github.com/%s/%s/%s/%d", parsed.Owner, parsed.Repo, seg, parsed.Number)
		}

		items = append(items, search.Item{
			FilePath: parsed.FilePath,
			Owner:    parsed.Owner,
			Repo:     parsed.Repo,
			Number:   parsed.Number,
			Type:     itemType,
			State:    strings.ToLower(parsed.State),
			Title:    parsed.Title,
			URL:      url,
		})

		return nil
	})

	s.Stop()

	return items, err
}

func discoverWithFilters(repo string) ([]search.Item, error) {
	filters := search.Filters{
		Repo:        repo,
		Issues:      rootIssues,
		PRs:         rootPRs,
		Discussions: rootDiscussions,
	}

	items, err := search.DiscoverLocalFiles(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to discover local files: %w", err)
	}

	return items, nil
}

func executeAction(cmd *cobra.Command, item *search.Item, action search.Action) error {
	p := output.NewPrinter(cmd)

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
		p.Print("Copied to clipboard")
		return nil

	case search.ActionPullFresh:
		return runPull(cmd, []string{item.FilePath})

	case search.ActionCancel:
		return nil

	default:
		return nil
	}
}

// itemOutput is the output structure for list items.
type itemOutput struct {
	Owner    string `json:"owner" yaml:"owner"`
	Repo     string `json:"repo" yaml:"repo"`
	Number   int    `json:"number" yaml:"number"`
	Type     string `json:"type" yaml:"type"`
	State    string `json:"state" yaml:"state"`
	Title    string `json:"title" yaml:"title"`
	URL      string `json:"url" yaml:"url"`
	FilePath string `json:"file_path" yaml:"file_path"`
}

func outputItems(p *output.Printer, items []search.Item) error {
	// Convert to output type
	out := make([]itemOutput, len(items))
	for i, item := range items {
		out[i] = itemOutput{
			Owner:    item.Owner,
			Repo:     item.Repo,
			Number:   item.Number,
			Type:     item.Type,
			State:    item.State,
			Title:    item.Title,
			URL:      item.URL,
			FilePath: item.FilePath,
		}
	}

	return output.List(p,
		[]string{"REPO", "NUMBER", "TYPE", "STATE", "TITLE", "PATH"},
		out,
		func(item itemOutput) []string {
			return []string{
				fmt.Sprintf("%s/%s", item.Owner, item.Repo),
				fmt.Sprintf("#%d", item.Number),
				item.Type,
				item.State,
				item.Title,
				item.FilePath,
			}
		},
	)
}
