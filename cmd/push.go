package cmd

import (
	"fmt"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/jackchuka/gh-md/internal/github"
	"github.com/jackchuka/gh-md/internal/output"
	"github.com/jackchuka/gh-md/internal/parser"
	"github.com/jackchuka/gh-md/internal/writer"
	"github.com/spf13/cobra"
)

var (
	pushForce  bool
	pushDryRun bool
)

var pushCmd = &cobra.Command{
	Use:   "push <file-path | url>",
	Short: "Push local changes back to GitHub",
	Long: `Push local markdown changes back to GitHub.

Supports pushing:
  - Title and body changes
  - State changes (open/closed) for issues and PRs
  - New comments
  - Edited comments

Examples:
  gh md push ~/.gh-md/owner/repo/issues/123.md
  gh md push https://github.com/owner/repo/issues/123
  gh md push --dry-run <file>`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Push even if remote has newer changes")
	pushCmd.Flags().BoolVar(&pushDryRun, "dry-run", false, "Show what would be pushed without making changes")
}

// changePlan represents all changes to be pushed.
type changePlan struct {
	titleBodyChanged bool
	stateChange      string // "", "close", or "reopen"
	newComments      []parser.ParsedComment
	editedComments   []parser.ParsedComment
}

func runPush(cmd *cobra.Command, args []string) error {
	p := output.NewPrinter(cmd)

	// Resolve input to file path
	filePath, err := parser.ResolveFilePath(args[0])
	if err != nil {
		return err
	}

	// Parse the markdown file
	parsed, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if parsed.ItemType == "" {
		return fmt.Errorf("could not determine item type from path: %s", filePath)
	}

	// Create GitHub client
	client, err := github.NewClient()
	if err != nil {
		return err
	}

	// Conflict check and fetch remote state
	s := newSpinner(cmd.ErrOrStderr(), "Checking for conflicts...")
	s.Start()

	remoteState, err := client.FetchRemoteState(parsed.ItemType, parsed.Owner, parsed.Repo, parsed.Number)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch remote state: %w", err)
	}

	// Fetch remote comments for diff comparison
	remoteComments, err := client.FetchComments(parsed.ItemType, parsed.Owner, parsed.Repo, parsed.Number)
	s.Stop()
	if err != nil {
		return fmt.Errorf("failed to fetch remote comments: %w", err)
	}

	if remoteState.UpdatedAt.After(parsed.Updated) {
		if !pushForce {
			p.Errorf("Conflict: remote has been updated since last pull\n")
			p.Errorf("  Local:  %s\n", output.FormatTime(&parsed.Updated, output.TimestampDisplay))
			p.Errorf("  Remote: %s\n", output.FormatTime(&remoteState.UpdatedAt, output.TimestampDisplay))
			p.Errorf("\nRun 'gh md pull' first, or use --force to override\n")
			return fmt.Errorf("conflict detected")
		}
		p.Errorf("Warning: overriding conflict (remote updated at %s)\n", output.FormatTime(&remoteState.UpdatedAt, output.TimestampDisplay))
	}

	// Build change plan
	plan := buildChangePlan(parsed, remoteState, remoteComments)

	// Dry run - show what would be pushed
	if pushDryRun {
		printDryRun(p, parsed, plan)
		return nil
	}

	// Check if there are any changes
	if !hasChanges(plan) {
		p.Printf("No changes to push for %s #%d\n", parsed.ItemType, parsed.Number)
		return nil
	}

	// Execute changes
	if err := executeChanges(p, client, parsed, plan, s); err != nil {
		return err
	}

	// Re-pull to sync timestamps
	s.Suffix = " Syncing local file..."
	s.Start()

	err = repullItem(client, parsed)
	s.Stop()
	if err != nil {
		p.Errorf("Warning: failed to sync local file: %v\n", err)
		p.Errorf("Run 'gh md pull' to sync manually\n")
	}

	return nil
}

func buildChangePlan(parsed *parser.ParsedFile, remoteState github.RemoteState, remoteComments []github.RemoteComment) changePlan {
	plan := changePlan{
		titleBodyChanged: true, // Always update title/body for now
	}

	// Check state change (only for issues and PRs)
	// Compare local state with remote state - only push if different
	if parsed.ItemType != github.ItemTypeDiscussion && parsed.State != "" && remoteState.State != "" {
		localState := strings.ToUpper(parsed.State)
		// Remote state is already uppercase (OPEN, CLOSED, MERGED)
		if localState != remoteState.State {
			switch localState {
			case "CLOSED":
				plan.stateChange = "close"
			case "OPEN":
				// Only reopen if not merged (merged PRs can't be reopened)
				if remoteState.State != "MERGED" {
					plan.stateChange = "reopen"
				}
			}
		}
	}

	// Build map of remote comments for comparison
	remoteMap := make(map[string]string)
	for _, rc := range remoteComments {
		remoteMap[rc.ID] = rc.Body
	}

	// Categorize local comments
	for _, c := range parsed.Comments {
		if c.ID == "" {
			// New comment (no ID)
			plan.newComments = append(plan.newComments, c)
		} else {
			// Existing comment - check if edited
			if remoteBody, ok := remoteMap[c.ID]; ok {
				if normalizeBody(c.Body) != normalizeBody(remoteBody) {
					plan.editedComments = append(plan.editedComments, c)
				}
			}
		}
	}

	return plan
}

func normalizeBody(body string) string {
	return strings.TrimSpace(body)
}

func hasChanges(plan changePlan) bool {
	return plan.titleBodyChanged || plan.stateChange != "" ||
		len(plan.newComments) > 0 || len(plan.editedComments) > 0
}

func printDryRun(p *output.Printer, parsed *parser.ParsedFile, plan changePlan) {
	p.Printf("Would push %s #%d:\n", parsed.ItemType, parsed.Number)
	p.Printf("  Title: %s\n", parsed.Title)
	p.Printf("  Body:  %d characters\n", len(parsed.Body))

	if plan.stateChange != "" {
		p.Printf("  State: %s\n", plan.stateChange)
	}

	if len(plan.newComments) > 0 {
		p.Printf("  New comments: %d\n", len(plan.newComments))
		for i, c := range plan.newComments {
			preview := c.Body
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			if c.ParentID != "" {
				p.Printf("    %d. (reply) %s\n", i+1, preview)
			} else {
				p.Printf("    %d. %s\n", i+1, preview)
			}
		}
	}

	if len(plan.editedComments) > 0 {
		p.Printf("  Edited comments: %d\n", len(plan.editedComments))
		for i, c := range plan.editedComments {
			p.Printf("    %d. %s\n", i+1, c.ID)
		}
	}
}

func executeChanges(p *output.Printer, client *github.Client, parsed *parser.ParsedFile, plan changePlan, s *spinner.Spinner) error {
	// 1. Update title/body
	s.Suffix = fmt.Sprintf(" Pushing %s #%d...", parsed.ItemType, parsed.Number)
	s.Start()

	var err error
	switch parsed.ItemType {
	case github.ItemTypeIssue:
		err = client.UpdateIssue(parsed.ID, parsed.Title, parsed.Body)
	case github.ItemTypePullRequest:
		err = client.UpdatePullRequest(parsed.ID, parsed.Title, parsed.Body)
	case github.ItemTypeDiscussion:
		err = client.UpdateDiscussion(parsed.ID, parsed.Title, parsed.Body)
	}

	s.Stop()
	if err != nil {
		return err
	}
	p.Printf("Pushed %s #%d\n", parsed.ItemType, parsed.Number)

	// 2. Update state (issues and PRs only)
	if plan.stateChange != "" && parsed.ItemType != github.ItemTypeDiscussion {
		s.Suffix = fmt.Sprintf(" Updating state to %s...", plan.stateChange)
		s.Start()

		switch parsed.ItemType {
		case github.ItemTypeIssue:
			if plan.stateChange == "close" {
				err = client.CloseIssue(parsed.ID)
			} else {
				err = client.ReopenIssue(parsed.ID)
			}
		case github.ItemTypePullRequest:
			if plan.stateChange == "close" {
				err = client.ClosePullRequest(parsed.ID)
			} else {
				err = client.ReopenPullRequest(parsed.ID)
			}
		}

		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to %s: %w", plan.stateChange, err)
		}
		p.Printf("State changed to %s\n", plan.stateChange)
	}

	// 3. Update existing comments
	for _, c := range plan.editedComments {
		s.Suffix = fmt.Sprintf(" Updating comment %s...", c.ID)
		s.Start()

		if parsed.ItemType == github.ItemTypeDiscussion {
			err = client.UpdateDiscussionComment(c.ID, c.Body)
		} else {
			err = client.UpdateIssueComment(c.ID, c.Body)
		}

		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to update comment %s: %w", c.ID, err)
		}
		p.Printf("Updated comment %s\n", c.ID)
	}

	// 4. Add new comments
	for _, c := range plan.newComments {
		s.Suffix = " Adding new comment..."
		s.Start()

		switch parsed.ItemType {
		case github.ItemTypeDiscussion:
			if c.ParentID != "" {
				// Reply to existing discussion comment
				err = client.AddDiscussionCommentReply(c.ParentID, c.Body)
			} else {
				// Top-level discussion comment
				err = client.AddDiscussionComment(parsed.ID, c.Body)
			}
		case github.ItemTypePullRequest:
			if c.ParentID != "" {
				// Reply to review thread
				err = client.AddReviewThreadReply(c.ParentID, c.Body)
			} else {
				// Regular PR comment
				err = client.AddComment(parsed.ID, c.Body)
			}
		default:
			// Issues use the same mutation for all comments
			err = client.AddComment(parsed.ID, c.Body)
		}

		s.Stop()
		if err != nil {
			return fmt.Errorf("failed to add comment: %w", err)
		}
		p.Printf("Added new comment\n")
	}

	return nil
}

func repullItem(client *github.Client, parsed *parser.ParsedFile) error {
	switch parsed.ItemType {
	case github.ItemTypeIssue:
		issue, err := client.FetchIssue(parsed.Owner, parsed.Repo, parsed.Number)
		if err != nil {
			return err
		}
		_, err = writer.WriteIssue(issue)
		return err

	case github.ItemTypePullRequest:
		pr, err := client.FetchPullRequest(parsed.Owner, parsed.Repo, parsed.Number)
		if err != nil {
			return err
		}
		_, err = writer.WritePullRequest(pr)
		return err

	case github.ItemTypeDiscussion:
		d, err := client.FetchDiscussion(parsed.Owner, parsed.Repo, parsed.Number)
		if err != nil {
			return err
		}
		_, err = writer.WriteDiscussion(d)
		return err

	default:
		return fmt.Errorf("unknown item type: %s", parsed.ItemType)
	}
}
