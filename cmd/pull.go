package cmd

import (
	"fmt"
	"time"

	"github.com/jackchuka/gh-md/internal/github"
	"github.com/jackchuka/gh-md/internal/meta"
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
)

var pullCmd = &cobra.Command{
	Use:   "pull <owner/repo | url | owner/repo/<type>/<number>>",
	Short: "Pull GitHub data to local markdown files",
	Long: `Pull issues, PRs, and discussions from GitHub and save them as local markdown files.

By default, all items (open and closed) are fetched for accurate state tracking.
Incremental sync is used automatically - only items updated since the last pull are fetched.
Single-item pulls (e.g., owner/repo/issues/123) always fetch regardless of state.

Examples:
  gh md pull owner/repo
  gh md pull owner/repo --issues --limit 10
  gh md pull owner/repo --open-only
  gh md pull owner/repo --full
  gh md pull https://github.com/owner/repo/issues/123
  gh md pull owner/repo/issues/123.md`,
	Args: cobra.ExactArgs(1),
	RunE: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)

	registerItemTypeFlags(pullCmd, &pullIssues, &pullPRs, &pullDiscussions, "Pull")
	pullCmd.Flags().IntVar(&pullLimit, "limit", 0, "Limit the number of items to pull (0 = no limit)")
	pullCmd.Flags().BoolVar(&pullOpenOnly, "open-only", false, "Fetch only open items (default fetches all states)")
	pullCmd.Flags().BoolVar(&pullFull, "full", false, "Full sync - ignore last sync timestamp")
}

func runPull(cmd *cobra.Command, args []string) error {
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

	// Load sync metadata
	md, err := meta.Load(input.Owner, input.Repo)
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
					input.Owner,
					input.Repo,
					github.ItemTypeIssue,
					func(progress github.ProgressFunc) ([]github.Issue, error) {
						return client.FetchIssues(input.Owner, input.Repo, pullLimit, pullOpenOnly, issuesSince, progress)
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
					input.Owner,
					input.Repo,
					github.ItemTypePullRequest,
					func(progress github.ProgressFunc) ([]github.PullRequest, error) {
						return client.FetchPullRequests(input.Owner, input.Repo, pullLimit, pullOpenOnly, pullsSince, progress)
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
					input.Owner,
					input.Repo,
					github.ItemTypeDiscussion,
					func(progress github.ProgressFunc) ([]github.Discussion, error) {
						return client.FetchDiscussions(input.Owner, input.Repo, pullLimit, pullOpenOnly, discussionsSince, progress)
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

	// Save sync timestamps on success
	if len(totalErrors) == 0 {
		if md.Sync == nil {
			md.Sync = &meta.SyncTimestamps{}
		}
		if pullAll || pullIssues {
			md.Sync.Issues = &syncStart
		}
		if pullAll || pullPRs {
			md.Sync.Pulls = &syncStart
		}
		if pullAll || pullDiscussions {
			md.Sync.Discussions = &syncStart
		}
		if err := meta.Save(input.Owner, input.Repo, md); err != nil {
			cmd.PrintErrf("Warning: failed to save sync metadata: %v\n", err)
		}
	}

	if len(totalErrors) > 0 {
		cmd.PrintErrln("\nSome errors occurred:")
		for _, e := range totalErrors {
			cmd.PrintErrf("  - %v\n", e)
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
	cmd.Printf("Wrote %s #%d to %s\n", writtenLabel, number(item), path)
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
		cmd.Printf("No %s found\n", plural)
		return nil
	}

	for i := range items {
		item := &items[i]
		if _, err := write(item); err != nil {
			return fmt.Errorf("failed to write %s #%d: %w", itemType.Display(), number(item), err)
		}
	}

	cmd.Printf("Wrote %d %s\n", len(items), plural)
	return nil
}
