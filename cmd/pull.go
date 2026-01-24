package cmd

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
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

	pullCmd.Flags().BoolVar(&pullIssues, "issues", false, "Pull only issues")
	pullCmd.Flags().BoolVar(&pullPRs, "prs", false, "Pull only pull requests")
	pullCmd.Flags().BoolVar(&pullDiscussions, "discussions", false, "Pull only discussions")
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

	if pullAll || pullIssues {
		if err := pullAllIssues(cmd, client, input.Owner, input.Repo); err != nil {
			totalErrors = append(totalErrors, fmt.Errorf("issues: %w", err))
		}
	}

	if pullAll || pullPRs {
		if err := pullAllPRs(cmd, client, input.Owner, input.Repo); err != nil {
			totalErrors = append(totalErrors, fmt.Errorf("pull requests: %w", err))
		}
	}

	if pullAll || pullDiscussions {
		if err := pullAllDiscussions(cmd, client, input.Owner, input.Repo); err != nil {
			totalErrors = append(totalErrors, fmt.Errorf("discussions: %w", err))
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
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = cmd.ErrOrStderr()
	s.Suffix = fmt.Sprintf(" Fetching %s #%d...", input.ItemType, input.Number)
	s.Start()

	switch input.ItemType {
	case github.ItemTypeIssue:
		issue, err := client.FetchIssue(input.Owner, input.Repo, input.Number)
		s.Stop()
		if err != nil {
			return err
		}
		path, err := writer.WriteIssue(issue)
		if err != nil {
			return err
		}
		cmd.Printf("Wrote issue #%d to %s\n", issue.Number, path)

	case github.ItemTypePullRequest:
		pr, err := client.FetchPullRequest(input.Owner, input.Repo, input.Number)
		s.Stop()
		if err != nil {
			return err
		}
		path, err := writer.WritePullRequest(pr)
		if err != nil {
			return err
		}
		cmd.Printf("Wrote PR #%d to %s\n", pr.Number, path)

	case github.ItemTypeDiscussion:
		d, err := client.FetchDiscussion(input.Owner, input.Repo, input.Number)
		s.Stop()
		if err != nil {
			return err
		}
		path, err := writer.WriteDiscussion(d)
		if err != nil {
			return err
		}
		cmd.Printf("Wrote discussion #%d to %s\n", d.Number, path)
	}

	return nil
}

func pullAllIssues(cmd *cobra.Command, client *github.Client, owner, repo string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = cmd.ErrOrStderr()
	s.Suffix = fmt.Sprintf(" Fetching issues from %s/%s...", owner, repo)
	s.Start()

	issues, err := client.FetchIssues(owner, repo, pullLimit)
	s.Stop()
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		cmd.Println("No issues found")
		return nil
	}

	for _, issue := range issues {
		if _, err := writer.WriteIssue(&issue); err != nil {
			return fmt.Errorf("failed to write issue #%d: %w", issue.Number, err)
		}
	}

	cmd.Printf("Wrote %d issues\n", len(issues))
	return nil
}

func pullAllPRs(cmd *cobra.Command, client *github.Client, owner, repo string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = cmd.ErrOrStderr()
	s.Suffix = fmt.Sprintf(" Fetching pull requests from %s/%s...", owner, repo)
	s.Start()

	prs, err := client.FetchPullRequests(owner, repo, pullLimit)
	s.Stop()
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		cmd.Println("No pull requests found")
		return nil
	}

	for _, pr := range prs {
		if _, err := writer.WritePullRequest(&pr); err != nil {
			return fmt.Errorf("failed to write PR #%d: %w", pr.Number, err)
		}
	}

	cmd.Printf("Wrote %d pull requests\n", len(prs))
	return nil
}

func pullAllDiscussions(cmd *cobra.Command, client *github.Client, owner, repo string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = cmd.ErrOrStderr()
	s.Suffix = fmt.Sprintf(" Fetching discussions from %s/%s...", owner, repo)
	s.Start()

	discussions, err := client.FetchDiscussions(owner, repo, pullLimit)
	s.Stop()
	if err != nil {
		return err
	}

	if len(discussions) == 0 {
		cmd.Println("No discussions found")
		return nil
	}

	for _, d := range discussions {
		if _, err := writer.WriteDiscussion(&d); err != nil {
			return fmt.Errorf("failed to write discussion #%d: %w", d.Number, err)
		}
	}

	cmd.Printf("Wrote %d discussions\n", len(discussions))
	return nil
}
