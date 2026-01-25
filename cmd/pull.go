package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-md/internal/github"
	"github.com/jackchuka/gh-md/internal/writer"
	"github.com/spf13/cobra"
)

var (
	pullIssues      bool
	pullPRs         bool
	pullDiscussions bool
	pullLimit       int
)

var pullCmd = &cobra.Command{
	Use:   "pull <owner/repo | url | owner/repo/<type>/<number>>",
	Short: "Pull GitHub data to local markdown files",
	Long: `Pull issues, PRs, and discussions from GitHub and save them as local markdown files.

Examples:
  gh md pull owner/repo
  gh md pull owner/repo --issues --limit 10
  gh md pull https://github.com/owner/repo/issues/123
  gh md pull owner/repo/issues/123.md`,
	Args: cobra.ExactArgs(1),
	RunE: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)

	registerItemTypeFlags(pullCmd, &pullIssues, &pullPRs, &pullDiscussions, "Pull")
	pullCmd.Flags().IntVar(&pullLimit, "limit", 0, "Limit the number of items to pull (0 = no limit)")
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

	// If no type flags are set, pull all types
	pullAll := !pullIssues && !pullPRs && !pullDiscussions

	var totalErrors []error

	handlers := []struct {
		enabled bool
		label   string
		run     func() error
	}{
		{
			enabled: pullAll || pullIssues,
			label:   "issues",
			run: func() error {
				return pullAllItems(
					cmd,
					input.Owner,
					input.Repo,
					"issues",
					"issue",
					func() ([]github.Issue, error) { return client.FetchIssues(input.Owner, input.Repo, pullLimit) },
					writer.WriteIssue,
					func(i *github.Issue) int { return i.Number },
				)
			},
		},
		{
			enabled: pullAll || pullPRs,
			label:   "pull requests",
			run: func() error {
				return pullAllItems(
					cmd,
					input.Owner,
					input.Repo,
					"pull requests",
					"PR",
					func() ([]github.PullRequest, error) {
						return client.FetchPullRequests(input.Owner, input.Repo, pullLimit)
					},
					writer.WritePullRequest,
					func(pr *github.PullRequest) int { return pr.Number },
				)
			},
		},
		{
			enabled: pullAll || pullDiscussions,
			label:   "discussions",
			run: func() error {
				return pullAllItems(
					cmd,
					input.Owner,
					input.Repo,
					"discussions",
					"discussion",
					func() ([]github.Discussion, error) {
						return client.FetchDiscussions(input.Owner, input.Repo, pullLimit)
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
			totalErrors = append(totalErrors, fmt.Errorf("%s: %w", h.label, err))
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
	fetchLabel string, // "issues", "pull requests", "discussions"
	writeLabel string, // "issue", "PR", "discussion"
	fetch func() ([]T, error),
	write func(*T) (string, error),
	number func(*T) int,
) error {
	s := newSpinner(cmd.ErrOrStderr(), fmt.Sprintf("Fetching %s from %s/%s...", fetchLabel, owner, repo))
	s.Start()
	items, err := fetch()
	s.Stop()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		cmd.Printf("No %s found\n", fetchLabel)
		return nil
	}

	for i := range items {
		item := &items[i]
		if _, err := write(item); err != nil {
			return fmt.Errorf("failed to write %s #%d: %w", writeLabel, number(item), err)
		}
	}

	cmd.Printf("Wrote %d %s\n", len(items), fetchLabel)
	return nil
}
