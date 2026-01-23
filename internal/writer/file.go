package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/github"
)

// WriteIssue writes an issue to the filesystem.
func WriteIssue(issue *github.Issue) (string, error) {
	dir, err := config.GetIssuesDir(issue.Owner, issue.Repo)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := IssueToMarkdown(issue)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.md", issue.Number))
	if err := writeFile(path, content); err != nil {
		return "", err
	}

	return path, nil
}

// WritePullRequest writes a PR to the filesystem.
func WritePullRequest(pr *github.PullRequest) (string, error) {
	dir, err := config.GetPullsDir(pr.Owner, pr.Repo)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := PullRequestToMarkdown(pr)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.md", pr.Number))
	if err := writeFile(path, content); err != nil {
		return "", err
	}

	return path, nil
}

// WriteDiscussion writes a discussion to the filesystem.
func WriteDiscussion(d *github.Discussion) (string, error) {
	dir, err := config.GetDiscussionsDir(d.Owner, d.Repo)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := DiscussionToMarkdown(d)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.md", d.Number))
	if err := writeFile(path, content); err != nil {
		return "", err
	}

	return path, nil
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
