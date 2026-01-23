package writer

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackchuka/gh-md/internal/github"
	"gopkg.in/yaml.v3"
)

// IssueFrontmatter represents the YAML frontmatter for an issue.
type IssueFrontmatter struct {
	ID         string    `yaml:"id"`
	URL        string    `yaml:"url"`
	Number     int       `yaml:"number"`
	Owner      string    `yaml:"owner"`
	Repo       string    `yaml:"repo"`
	Title      string    `yaml:"title"`
	State      string    `yaml:"state"`
	Labels     []string  `yaml:"labels,omitempty"`
	Assignees  []string  `yaml:"assignees,omitempty"`
	Created    time.Time `yaml:"created"`
	Updated    time.Time `yaml:"updated"`
	LastPulled time.Time `yaml:"last_pulled"`
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
		Labels:     issue.Labels,
		Assignees:  issue.Assignees,
		Created:    issue.CreatedAt,
		Updated:    issue.UpdatedAt,
		LastPulled: time.Now().UTC(),
	}

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString("# ")
	sb.WriteString(issue.Title)
	sb.WriteString("\n\n")
	sb.WriteString(issue.Body)
	sb.WriteString("\n")

	if len(issue.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		sb.WriteString("<!-- gh-md:comments -->\n\n")

		for _, c := range issue.Comments {
			writeComment(&sb, c)
		}
	}

	return sb.String(), nil
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

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString("# ")
	sb.WriteString(pr.Title)
	sb.WriteString("\n\n")
	sb.WriteString(pr.Body)
	sb.WriteString("\n")

	if len(pr.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		sb.WriteString("<!-- gh-md:comments -->\n\n")

		for _, c := range pr.Comments {
			writeComment(&sb, c)
		}
	}

	return sb.String(), nil
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

	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmBytes)
	sb.WriteString("---\n\n")
	sb.WriteString("# ")
	sb.WriteString(d.Title)
	sb.WriteString("\n\n")
	sb.WriteString(d.Body)
	sb.WriteString("\n")

	if len(d.Comments) > 0 {
		sb.WriteString("\n---\n\n")
		sb.WriteString("<!-- gh-md:comments -->\n\n")

		for _, c := range d.Comments {
			writeDiscussionComment(&sb, c, 0)
		}
	}

	return sb.String(), nil
}

func writeComment(sb *strings.Builder, c github.Comment) {
	sb.WriteString("<!-- gh-md:comment-meta\n")
	fmt.Fprintf(sb, "id: %s\n", c.ID)
	fmt.Fprintf(sb, "author: %s\n", c.Author)
	fmt.Fprintf(sb, "created: %s\n", c.CreatedAt.Format(time.RFC3339))
	sb.WriteString("-->\n\n")
	fmt.Fprintf(sb, "### @%s (%s)\n\n", c.Author, c.CreatedAt.Format("2006-01-02"))
	sb.WriteString(c.Body)
	sb.WriteString("\n\n")
}

func writeDiscussionComment(sb *strings.Builder, c github.DiscussionComment, depth int) {
	indent := strings.Repeat("  ", depth)

	sb.WriteString(indent)
	sb.WriteString("<!-- gh-md:comment-meta\n")
	sb.WriteString(indent)
	fmt.Fprintf(sb, "id: %s\n", c.ID)
	sb.WriteString(indent)
	fmt.Fprintf(sb, "author: %s\n", c.Author)
	sb.WriteString(indent)
	fmt.Fprintf(sb, "created: %s\n", c.CreatedAt.Format(time.RFC3339))
	sb.WriteString(indent)
	sb.WriteString("-->\n\n")

	heading := "###"
	if depth > 0 {
		heading = "####"
	}
	sb.WriteString(indent)
	fmt.Fprintf(sb, "%s @%s (%s)\n\n", heading, c.Author, c.CreatedAt.Format("2006-01-02"))

	// Indent body lines
	for _, line := range strings.Split(c.Body, "\n") {
		sb.WriteString(indent)
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for _, reply := range c.Replies {
		writeDiscussionComment(sb, reply, depth+1)
	}
}
