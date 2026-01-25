package writer

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackchuka/gh-md/internal/github"
	"gopkg.in/yaml.v3"
)

// buildMarkdownWithFrontmatter creates a markdown document with YAML frontmatter.
func buildMarkdownWithFrontmatter(fm interface{}, title, body string) (*strings.Builder, error) {
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString("<!-- gh-md:content -->\n")
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n\n")
	sb.WriteString(body)
	sb.WriteString("\n<!-- /gh-md:content -->\n")

	return &sb, nil
}

// finishMarkdown adds the new-comment marker and returns the final string.
func finishMarkdown(sb *strings.Builder) string {
	sb.WriteString("\n<!-- gh-md:new-comment -->\n\n<!-- /gh-md:new-comment -->\n")
	return sb.String()
}

// writeCommentHeader writes the common comment metadata block.
func writeCommentHeader(sb *strings.Builder, tagName, id, author string, createdAt time.Time) {
	fmt.Fprintf(sb, "<!-- gh-md:%s\n", tagName)
	fmt.Fprintf(sb, "id: %s\n", id)
	fmt.Fprintf(sb, "author: %s\n", author)
	fmt.Fprintf(sb, "created: %s\n", createdAt.Format(time.RFC3339))
	sb.WriteString("-->\n")
}

// writeCommentBody writes the author heading and body with closing tag.
func writeCommentBody(sb *strings.Builder, tagName, author, body string, createdAt time.Time, headingLevel string) {
	fmt.Fprintf(sb, "%s @%s (%s)\n\n", headingLevel, author, createdAt.Format("2006-01-02"))
	sb.WriteString(body)
	fmt.Fprintf(sb, "\n<!-- /gh-md:%s -->\n\n", tagName)
}

// IssueReferenceFrontmatter represents a reference to a parent or child issue in frontmatter.
type IssueReferenceFrontmatter struct {
	Number int    `yaml:"number"`
	Title  string `yaml:"title"`
	URL    string `yaml:"url"`
	State  string `yaml:"state"`
	Owner  string `yaml:"owner,omitempty"` // Only if cross-repo
	Repo   string `yaml:"repo,omitempty"`  // Only if cross-repo
}

// SubIssuesSummaryFrontmatter provides statistics about sub-issues in frontmatter.
type SubIssuesSummaryFrontmatter struct {
	Total           int `yaml:"total"`
	Completed       int `yaml:"completed"`
	PercentComplete int `yaml:"percent_complete"`
}

// IssueFrontmatter represents the YAML frontmatter for an issue.
type IssueFrontmatter struct {
	ID               string                       `yaml:"id"`
	URL              string                       `yaml:"url"`
	Number           int                          `yaml:"number"`
	Owner            string                       `yaml:"owner"`
	Repo             string                       `yaml:"repo"`
	Title            string                       `yaml:"title"`
	State            string                       `yaml:"state"`
	Author           string                       `yaml:"author,omitempty"`
	Labels           []string                     `yaml:"labels,omitempty"`
	Assignees        []string                     `yaml:"assignees,omitempty"`
	Created          time.Time                    `yaml:"created"`
	Updated          time.Time                    `yaml:"updated"`
	LastPulled       time.Time                    `yaml:"last_pulled"`
	Parent           *IssueReferenceFrontmatter   `yaml:"parent,omitempty"`
	Children         []IssueReferenceFrontmatter  `yaml:"children,omitempty"`
	SubIssuesSummary *SubIssuesSummaryFrontmatter `yaml:"sub_issues_summary,omitempty"`
}

// PullRequestFrontmatter represents the YAML frontmatter for a PR.
type PullRequestFrontmatter struct {
	ID          string    `yaml:"id"`
	URL         string    `yaml:"url"`
	Number      int       `yaml:"number"`
	Owner       string    `yaml:"owner"`
	Repo        string    `yaml:"repo"`
	Title       string    `yaml:"title"`
	State       string    `yaml:"state"`
	Author      string    `yaml:"author,omitempty"`
	Draft       bool      `yaml:"draft,omitempty"`
	Labels      []string  `yaml:"labels,omitempty"`
	Assignees   []string  `yaml:"assignees,omitempty"`
	HeadRef     string    `yaml:"head_ref"`
	BaseRef     string    `yaml:"base_ref"`
	MergeCommit string    `yaml:"merge_commit,omitempty"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Merged      time.Time `yaml:"merged,omitempty"`
	LastPulled  time.Time `yaml:"last_pulled"`
}

// DiscussionFrontmatter represents the YAML frontmatter for a discussion.
type DiscussionFrontmatter struct {
	ID         string    `yaml:"id"`
	URL        string    `yaml:"url"`
	Number     int       `yaml:"number"`
	Owner      string    `yaml:"owner"`
	Repo       string    `yaml:"repo"`
	Title      string    `yaml:"title"`
	Category   string    `yaml:"category"`
	Author     string    `yaml:"author"`
	AnswerID   string    `yaml:"answer_id,omitempty"`
	Locked     bool      `yaml:"locked,omitempty"`
	Created    time.Time `yaml:"created"`
	Updated    time.Time `yaml:"updated"`
	LastPulled time.Time `yaml:"last_pulled"`
}

// IssueToMarkdown converts an issue to markdown with YAML frontmatter.
func IssueToMarkdown(issue *github.Issue) (string, error) {
	fm := IssueFrontmatter{
		ID:         issue.ID,
		URL:        issue.URL,
		Number:     issue.Number,
		Owner:      issue.Owner,
		Repo:       issue.Repo,
		Title:      issue.Title,
		State:      issue.State,
		Author:     issue.Author,
		Labels:     issue.Labels,
		Assignees:  issue.Assignees,
		Created:    issue.CreatedAt,
		Updated:    issue.UpdatedAt,
		LastPulled: time.Now().UTC(),
	}

	// Convert parent issue reference
	if issue.Parent != nil {
		fm.Parent = &IssueReferenceFrontmatter{
			Number: issue.Parent.Number,
			Title:  issue.Parent.Title,
			URL:    issue.Parent.URL,
			State:  issue.Parent.State,
		}
		// Only include owner/repo if cross-repo
		if issue.Parent.Owner != issue.Owner || issue.Parent.Repo != issue.Repo {
			fm.Parent.Owner = issue.Parent.Owner
			fm.Parent.Repo = issue.Parent.Repo
		}
	}

	// Convert children (sub-issues)
	if len(issue.Children) > 0 {
		fm.Children = make([]IssueReferenceFrontmatter, 0, len(issue.Children))
		for _, child := range issue.Children {
			childFm := IssueReferenceFrontmatter{
				Number: child.Number,
				Title:  child.Title,
				URL:    child.URL,
				State:  child.State,
			}
			// Only include owner/repo if cross-repo
			if child.Owner != issue.Owner || child.Repo != issue.Repo {
				childFm.Owner = child.Owner
				childFm.Repo = child.Repo
			}
			fm.Children = append(fm.Children, childFm)
		}
	}

	// Convert sub-issues summary
	if issue.SubIssuesSummary != nil {
		fm.SubIssuesSummary = &SubIssuesSummaryFrontmatter{
			Total:           issue.SubIssuesSummary.Total,
			Completed:       issue.SubIssuesSummary.Completed,
			PercentComplete: issue.SubIssuesSummary.PercentComplete,
		}
	}

	sb, err := buildMarkdownWithFrontmatter(fm, issue.Title, issue.Body)
	if err != nil {
		return "", err
	}

	if len(issue.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		for _, c := range issue.Comments {
			writeComment(sb, c)
		}
	}

	return finishMarkdown(sb), nil
}

// PullRequestToMarkdown converts a PR to markdown with YAML frontmatter.
func PullRequestToMarkdown(pr *github.PullRequest) (string, error) {
	fm := PullRequestFrontmatter{
		ID:          pr.ID,
		URL:         pr.URL,
		Number:      pr.Number,
		Owner:       pr.Owner,
		Repo:        pr.Repo,
		Title:       pr.Title,
		State:       pr.State,
		Author:      pr.Author,
		Draft:       pr.Draft,
		Labels:      pr.Labels,
		Assignees:   pr.Assignees,
		HeadRef:     pr.HeadRef,
		BaseRef:     pr.BaseRef,
		MergeCommit: pr.MergeCommit,
		Created:     pr.CreatedAt,
		Updated:     pr.UpdatedAt,
		LastPulled:  time.Now().UTC(),
	}

	if !pr.MergedAt.IsZero() {
		fm.Merged = pr.MergedAt
	}

	sb, err := buildMarkdownWithFrontmatter(fm, pr.Title, pr.Body)
	if err != nil {
		return "", err
	}

	if len(pr.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		for _, c := range pr.Comments {
			writeComment(sb, c)
		}
	}

	if len(pr.ReviewThreads) > 0 {
		sb.WriteString("\n## Review Threads\n\n")
		for _, thread := range pr.ReviewThreads {
			writeReviewThread(sb, thread)
		}
	}

	return finishMarkdown(sb), nil
}

// DiscussionToMarkdown converts a discussion to markdown with YAML frontmatter.
func DiscussionToMarkdown(d *github.Discussion) (string, error) {
	fm := DiscussionFrontmatter{
		ID:         d.ID,
		URL:        d.URL,
		Number:     d.Number,
		Owner:      d.Owner,
		Repo:       d.Repo,
		Title:      d.Title,
		Category:   d.Category,
		Author:     d.Author,
		AnswerID:   d.AnswerID,
		Locked:     d.Locked,
		Created:    d.CreatedAt,
		Updated:    d.UpdatedAt,
		LastPulled: time.Now().UTC(),
	}

	sb, err := buildMarkdownWithFrontmatter(fm, d.Title, d.Body)
	if err != nil {
		return "", err
	}

	if len(d.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		for _, c := range d.Comments {
			writeDiscussionComment(sb, c, "")
		}
	}

	return finishMarkdown(sb), nil
}

func writeComment(sb *strings.Builder, c github.Comment) {
	writeCommentHeader(sb, "comment", c.ID, c.Author, c.CreatedAt)
	writeCommentBody(sb, "comment", c.Author, c.Body, c.CreatedAt, "###")
}

func writeReviewThread(sb *strings.Builder, thread github.ReviewThread) {
	// Thread header
	resolved := ""
	if thread.IsResolved {
		resolved = " (resolved)"
	}
	outdated := ""
	if thread.IsOutdated {
		outdated = " (outdated)"
	}
	sb.WriteString("<!-- gh-md:review-thread\n")
	fmt.Fprintf(sb, "id: %s\n", thread.ID)
	fmt.Fprintf(sb, "path: %s\n", thread.Path)
	fmt.Fprintf(sb, "line: %d\n", thread.Line)
	sb.WriteString("-->\n")
	fmt.Fprintf(sb, "### `%s:%d`%s%s\n\n", thread.Path, thread.Line, resolved, outdated)

	// Write each comment in the thread
	for _, c := range thread.Comments {
		sb.WriteString("<!-- gh-md:review-comment\n")
		fmt.Fprintf(sb, "id: %s\n", c.ID)
		fmt.Fprintf(sb, "author: %s\n", c.Author)
		fmt.Fprintf(sb, "created: %s\n", c.CreatedAt.Format(time.RFC3339))
		sb.WriteString("-->\n")
		fmt.Fprintf(sb, "#### @%s (%s)\n\n", c.Author, c.CreatedAt.Format("2006-01-02"))
		sb.WriteString(c.Body)
		sb.WriteString("\n<!-- /gh-md:review-comment -->\n\n")
	}

	// Reply marker for this thread
	fmt.Fprintf(sb, "<!-- gh-md:new-comment reply_to: %s -->\n\n<!-- /gh-md:new-comment -->\n\n", thread.ID)

	sb.WriteString("<!-- /gh-md:review-thread -->\n\n")
}

func writeDiscussionComment(sb *strings.Builder, c github.DiscussionComment, parentID string) {
	sb.WriteString("<!-- gh-md:comment\n")
	fmt.Fprintf(sb, "id: %s\n", c.ID)
	fmt.Fprintf(sb, "author: %s\n", c.Author)
	if parentID != "" {
		fmt.Fprintf(sb, "parent: %s\n", parentID)
	}
	fmt.Fprintf(sb, "created: %s\n", c.CreatedAt.Format(time.RFC3339))
	sb.WriteString("-->\n")

	heading := "###"
	if parentID != "" {
		heading = "####"
	}
	fmt.Fprintf(sb, "%s @%s (%s)\n\n", heading, c.Author, c.CreatedAt.Format("2006-01-02"))

	sb.WriteString(c.Body)
	sb.WriteString("\n<!-- /gh-md:comment -->\n\n")

	// Add new reply marker for this comment
	fmt.Fprintf(sb, "<!-- gh-md:new-comment reply_to: %s -->\n\n<!-- /gh-md:new-comment -->\n\n", c.ID)

	for _, reply := range c.Replies {
		writeDiscussionComment(sb, reply, c.ID)
	}
}
