package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/github"
)

// Item is an interface for items that can be written to markdown.
type Item interface {
	GetOwner() string
	GetRepo() string
	GetNumber() int
}

// writeItemToFile is a generic helper for writing items to markdown files.
func writeItemToFile(
	item Item,
	getDirFunc func(owner, repo string) (string, error),
	toMarkdownFunc func() (string, error),
) (string, error) {
	dir, err := getDirFunc(item.GetOwner(), item.GetRepo())
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := toMarkdownFunc()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.md", item.GetNumber()))
	if err := writeFile(path, content); err != nil {
		return "", err
	}

	return path, nil
}

// WriteIssue writes an issue to the filesystem.
func WriteIssue(issue *github.Issue) (string, error) {
	return writeItemToFile(
		issue,
		config.GetIssuesDir,
		func() (string, error) { return IssueToMarkdown(issue) },
	)
}

// WritePullRequest writes a PR to the filesystem.
func WritePullRequest(pr *github.PullRequest) (string, error) {
	return writeItemToFile(
		pr,
		config.GetPullsDir,
		func() (string, error) { return PullRequestToMarkdown(pr) },
	)
}

// WriteDiscussion writes a discussion to the filesystem.
func WriteDiscussion(d *github.Discussion) (string, error) {
	return writeItemToFile(
		d,
		config.GetDiscussionsDir,
		func() (string, error) { return DiscussionToMarkdown(d) },
	)
}

// writeFile writes content to a file atomically by writing to a temp file first.
func writeFile(path, content string) error {
	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rename to final path (atomic on most filesystems)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
