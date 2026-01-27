package cmd

import (
	"fmt"
	"time"

	"github.com/jackchuka/gh-md/internal/discovery"
	"github.com/jackchuka/gh-md/internal/gitcontext"
	"github.com/jackchuka/gh-md/internal/github"
	"github.com/jackchuka/gh-md/internal/meta"
	"github.com/jackchuka/gh-md/internal/output"
	"github.com/jackchuka/gh-md/internal/writer"
	"github.com/spf13/cobra"
)

var (
	pullIssues      bool
	pullPRs         bool
	pullDiscussions bool
	pullLimit       int
	pullOpenOnly    bool
	pullFull        bool
	pullAllRepos    bool
)

var pullCmd = &cobra.Command{
	Use:   "pull [owner/repo | url | owner/repo/<type>/<number>]",
	Short: "Pull GitHub data to local markdown files",
	Long: `Pull issues, PRs, and discussions from GitHub and save them as local markdown files.

When run without arguments inside a git repository:
  - On a branch with an open PR: pulls that PR with all review comments
  - On main/master: pulls all items for the current repository
  - On a branch without a PR: shows an error with suggestions

When run with explicit arguments, pulls the specified items.

By default, all items (open and closed) are fetched for accurate state tracking.
Incremental sync is used automatically - only items updated since the last pull are fetched.
Single-item pulls (e.g., owner/repo/issues/123) always fetch regardless of state.

Examples:
  gh md pull                           # Smart pull based on current git context
  gh md pull owner/repo
  gh md pull owner/repo --issues --limit 10
  gh md pull owner/repo --open-only
  gh md pull owner/repo --full
  gh md pull --all
  gh md pull https://github.com/owner/repo/issues/123
  gh md pull owner/repo/issues/123.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)

	registerItemTypeFlags(pullCmd, &pullIssues, &pullPRs, &pullDiscussions, "Pull")
	pullCmd.Flags().IntVar(&pullLimit, "limit", 0, "Limit the number of items to pull (0 = no limit)")
	pullCmd.Flags().BoolVar(&pullOpenOnly, "open-only", false, "Fetch only open items (default fetches all states)")
	pullCmd.Flags().BoolVar(&pullFull, "full", false, "Full sync - ignore last sync timestamp")
	pullCmd.Flags().BoolVar(&pullAllRepos, "all", false, "Pull all managed repositories")
}

func runPull(cmd *cobra.Command, args []string) error {
	// Handle --all flag
	if pullAllRepos {
		if len(args) > 0 {
			return fmt.Errorf("--all flag cannot be used with a specific repository")
		}
		return runPullAll(cmd)
	}

	if len(args) == 0 {
		return runSmartPull(cmd)
	}

	input, err := github.ParseInput(args[0])
	if err != nil {
		return err
	}

	client, err := github.NewClient()
	if err != nil {
		return err
	}

	// If a specific item was requested via URL
	if input.Number > 0 {
		return pullSingleItem(cmd, client, input)
	}

	return pullRepo(cmd, client, input.Owner, input.Repo)
}

func runSmartPull(cmd *cobra.Command) error {
	p := output.NewPrinter(cmd)

	// Detect git context
	ctx, err := gitcontext.Detect()
	if err != nil {
		return fmt.Errorf("repository argument required (or use --all to pull all managed repos)\n\n%w", err)
	}

	p.Printf("Detected repository: %s/%s (branch: %s)\n", ctx.Owner, ctx.Repo, ctx.Branch)

	// Resolve what to pull
	result, err := ctx.Resolve()
	if err != nil {
		return err
	}

	client, err := github.NewClient()
	if err != nil {
		return err
	}

	if result.PRNumber > 0 {
		// Pull specific PR with reviews
		p.Printf("Pulling PR #%d...\n", result.PRNumber)
		input := &github.ParsedInput{
			Owner:    result.Owner,
			Repo:     result.Repo,
			Number:   result.PRNumber,
			ItemType: github.ItemTypePullRequest,
		}
		return pullSingleItem(cmd, client, input)
	}

	// On default branch - pull all items for this repo
	p.Printf("On default branch - pulling all items for %s/%s\n", result.Owner, result.Repo)
	return pullRepo(cmd, client, result.Owner, result.Repo)
}

func runPullAll(cmd *cobra.Command) error {
	p := output.NewPrinter(cmd)

	repos, err := discovery.DiscoverManagedRepos()
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(repos) == 0 {
		p.Print("No managed repositories found.")
		return nil
	}

	p.Printf("Pulling %d repositories...\n", len(repos))

	client, err := github.NewClient()
	if err != nil {
		return err
	}

	var errors []error
	for i, repo := range repos {
		p.Printf("[%d/%d] %s\n", i+1, len(repos), repo.Slug())

		if err := pullRepo(cmd, client, repo.Owner, repo.Repo); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", repo.Slug(), err))
			p.Errorf("  Error: %v\n", err)
		}
	}

	p.Printf("\nCompleted: %d/%d repositories\n", len(repos)-len(errors), len(repos))

	if len(errors) > 0 {
		return fmt.Errorf("%d repositories failed to update", len(errors))
	}

	return nil
}

func pullRepo(cmd *cobra.Command, client *github.Client, owner, repo string) error {
	// Load sync metadata
	md, err := meta.Load(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to load sync metadata: %w", err)
	}

	// Get since timestamps (nil if --full or no prior sync)
	var issuesSince, pullsSince, discussionsSince *time.Time
	if !pullFull && md.Sync != nil {
		issuesSince = md.Sync.Issues
		pullsSince = md.Sync.Pulls
		discussionsSince = md.Sync.Discussions
	}

	// Track sync start time for saving later
	syncStart := time.Now()

	// If no type flags are set, pull all types
	pullAll := !pullIssues && !pullPRs && !pullDiscussions

	var totalErrors []error

	handlers := []struct {
		enabled  bool
		itemType github.ItemType
		run      func() error
	}{
		{
			enabled:  pullAll || pullIssues,
			itemType: github.ItemTypeIssue,
			run: func() error {
				return pullAllItems(
					cmd,
					owner,
					repo,
					github.ItemTypeIssue,
					func(progress github.ProgressFunc) ([]github.Issue, error) {
						return client.FetchIssues(owner, repo, pullLimit, pullOpenOnly, issuesSince, progress)
					},
					writer.WriteIssue,
					func(i *github.Issue) int { return i.Number },
				)
			},
		},
		{
			enabled:  pullAll || pullPRs,
			itemType: github.ItemTypePullRequest,
			run: func() error {
				return pullAllItems(
					cmd,
					owner,
					repo,
					github.ItemTypePullRequest,
					func(progress github.ProgressFunc) ([]github.PullRequest, error) {
						return client.FetchPullRequests(owner, repo, pullLimit, pullOpenOnly, pullsSince, progress)
					},
					writer.WritePullRequest,
					func(pr *github.PullRequest) int { return pr.Number },
				)
			},
		},
		{
			enabled:  pullAll || pullDiscussions,
			itemType: github.ItemTypeDiscussion,
			run: func() error {
				return pullAllItems(
					cmd,
					owner,
					repo,
					github.ItemTypeDiscussion,
					func(progress github.ProgressFunc) ([]github.Discussion, error) {
						return client.FetchDiscussions(owner, repo, pullLimit, pullOpenOnly, discussionsSince, progress)
					},
					writer.WriteDiscussion,
					func(d *github.Discussion) int { return d.Number },
				)
			},
		},
	}

	for _, h := range handlers {
		if !h.enabled {
			continue
		}
		if err := h.run(); err != nil {
			totalErrors = append(totalErrors, fmt.Errorf("%s: %w", h.itemType.DisplayPlural(), err))
		}
	}

	p := output.NewPrinter(cmd)

	// Save sync timestamps on success
	if len(totalErrors) == 0 {
		if md.Sync == nil {
			md.Sync = &meta.SyncTimestamps{}
		}
		// Save previous timestamps for --new flag
		if pullAll || pullIssues {
			md.Sync.PrevIssues = md.Sync.Issues
			md.Sync.Issues = &syncStart
		}
		if pullAll || pullPRs {
			md.Sync.PrevPulls = md.Sync.Pulls
			md.Sync.Pulls = &syncStart
		}
		if pullAll || pullDiscussions {
			md.Sync.PrevDiscussions = md.Sync.Discussions
			md.Sync.Discussions = &syncStart
		}
		if err := meta.Save(owner, repo, md); err != nil {
			p.Errorf("Warning: failed to save sync metadata: %v\n", err)
		}
	}

	if len(totalErrors) > 0 {
		p.Errorf("  Some errors occurred:\n")
		for _, e := range totalErrors {
			p.Errorf("    - %v\n", e)
		}
		return fmt.Errorf("pull completed with %d error(s)", len(totalErrors))
	}

	return nil
}

func pullSingleItem(cmd *cobra.Command, client *github.Client, input *github.ParsedInput) error {
	handlers := map[github.ItemType]func() error{
		github.ItemTypeIssue: func() error {
			return pullSingle(
				cmd,
				input,
				func() (*github.Issue, error) { return client.FetchIssue(input.Owner, input.Repo, input.Number) },
				writer.WriteIssue,
				"issue",
				func(i *github.Issue) int { return i.Number },
			)
		},
		github.ItemTypePullRequest: func() error {
			return pullSingle(
				cmd,
				input,
				func() (*github.PullRequest, error) {
					return client.FetchPullRequest(input.Owner, input.Repo, input.Number)
				},
				writer.WritePullRequest,
				"PR",
				func(pr *github.PullRequest) int { return pr.Number },
			)
		},
		github.ItemTypeDiscussion: func() error {
			return pullSingle(
				cmd,
				input,
				func() (*github.Discussion, error) {
					return client.FetchDiscussion(input.Owner, input.Repo, input.Number)
				},
				writer.WriteDiscussion,
				"discussion",
				func(d *github.Discussion) int { return d.Number },
			)
		},
	}

	handler, ok := handlers[input.ItemType]
	if !ok {
		return fmt.Errorf("unsupported item type: %s", input.ItemType)
	}
	return handler()
}

func pullSingle[T any](
	cmd *cobra.Command,
	input *github.ParsedInput,
	fetch func() (*T, error),
	write func(*T) (string, error),
	writtenLabel string,
	number func(*T) int,
) error {
	p := output.NewPrinter(cmd)
	s := newSpinner(cmd.ErrOrStderr(), fmt.Sprintf("Fetching %s #%d...", input.ItemType, input.Number))
	s.Start()
	item, err := fetch()
	s.Stop()
	if err != nil {
		return err
	}
	path, err := write(item)
	if err != nil {
		return err
	}
	p.Printf("Wrote %s #%d to %s\n", writtenLabel, number(item), path)
	return nil
}

func pullAllItems[T any](
	cmd *cobra.Command,
	owner, repo string,
	itemType github.ItemType,
	fetch func(github.ProgressFunc) ([]T, error),
	write func(*T) (string, error),
	number func(*T) int,
) error {
	p := output.NewPrinter(cmd)
	plural := itemType.DisplayPlural()
	s := newSpinner(cmd.ErrOrStderr(), fmt.Sprintf("Fetching %s from %s/%s...", plural, owner, repo))
	s.Start()

	progress := func(fetched int) {
		s.Suffix = fmt.Sprintf(" Fetching %s from %s/%s... (%d)", plural, owner, repo, fetched)
	}

	items, err := fetch(progress)
	s.Stop()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		p.Printf("No %s found.\n", plural)
		return nil
	}

	for i := range items {
		item := &items[i]
		if _, err := write(item); err != nil {
			return fmt.Errorf("failed to write %s #%d: %w", itemType.Display(), number(item), err)
		}
	}

	p.Printf("Wrote %d %s\n", len(items), plural)
	return nil
}
