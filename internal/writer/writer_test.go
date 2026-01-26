package writer

import (
	"strings"
	"testing"
	"time"

	"github.com/jackchuka/gh-md/internal/github"
)

func TestIssueToMarkdown(t *testing.T) {
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		issue      *github.Issue
		wantParts  []string
		wantNoPart []string
	}{
		{
			name: "basic issue",
			issue: &github.Issue{
				ID:        "I_123",
				URL:       "https://github.com/owner/repo/issues/1",
				Number:    1,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Test Issue",
				Body:      "This is the body",
				State:     "open",
				Author:    "octocat",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"---",
				"id: I_123",
				"number: 1",
				"owner: owner",
				"repo: repo",
				"state: open",
				"# Test Issue",
				"This is the body",
				"<!-- gh-md:content -->",
				"<!-- /gh-md:content -->",
				"<!-- gh-md:new-comment -->",
				"<!-- /gh-md:new-comment -->",
			},
		},
		{
			name: "issue with labels and assignees",
			issue: &github.Issue{
				ID:        "I_456",
				URL:       "https://github.com/owner/repo/issues/2",
				Number:    2,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Bug Report",
				Body:      "Found a bug",
				State:     "open",
				Author:    "user1",
				Labels:    []string{"bug", "critical"},
				Assignees: []string{"dev1", "dev2"},
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"labels:",
				"- bug",
				"- critical",
				"assignees:",
				"- dev1",
				"- dev2",
			},
		},
		{
			name: "issue with comments",
			issue: &github.Issue{
				ID:        "I_789",
				URL:       "https://github.com/owner/repo/issues/3",
				Number:    3,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Issue with Comments",
				Body:      "Main body",
				State:     "open",
				Author:    "author",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				Comments: []github.Comment{
					{
						ID:        "IC_001",
						Author:    "commenter",
						Body:      "This is a comment",
						CreatedAt: baseTime,
					},
				},
			},
			wantParts: []string{
				"<!-- gh-md:comment",
				"id: IC_001",
				"author: commenter",
				"This is a comment",
				"<!-- /gh-md:comment -->",
				"### @commenter",
			},
		},
		{
			name: "issue with parent",
			issue: &github.Issue{
				ID:        "I_child",
				URL:       "https://github.com/owner/repo/issues/10",
				Number:    10,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Child Issue",
				Body:      "Sub-task",
				State:     "open",
				Author:    "author",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				Parent: &github.IssueReference{
					Number: 5,
					Title:  "Parent Issue",
					URL:    "https://github.com/owner/repo/issues/5",
					State:  "open",
					Owner:  "owner",
					Repo:   "repo",
				},
			},
			wantParts: []string{
				"parent:",
				"number: 5",
				"title: Parent Issue",
			},
			wantNoPart: []string{
				// Same repo parent should not include owner/repo
				"parent:\n    owner:",
			},
		},
		{
			name: "issue with children",
			issue: &github.Issue{
				ID:        "I_parent",
				URL:       "https://github.com/owner/repo/issues/5",
				Number:    5,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Parent Issue",
				Body:      "Main task",
				State:     "open",
				Author:    "author",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				Children: []github.IssueReference{
					{Number: 10, Title: "Child 1", State: "open", Owner: "owner", Repo: "repo"},
					{Number: 11, Title: "Child 2", State: "closed", Owner: "owner", Repo: "repo"},
				},
				SubIssuesSummary: &github.SubIssuesSummary{
					Total:           2,
					Completed:       1,
					PercentComplete: 50,
				},
			},
			wantParts: []string{
				"children:",
				"- number: 10",
				"- number: 11",
				"sub_issues_summary:",
				"total: 2",
				"completed: 1",
				"percent_complete: 50",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IssueToMarkdown(tt.issue)
			if err != nil {
				t.Fatalf("IssueToMarkdown() error = %v", err)
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("IssueToMarkdown() missing %q\nGot:\n%s", part, got)
				}
			}

			for _, noPart := range tt.wantNoPart {
				if strings.Contains(got, noPart) {
					t.Errorf("IssueToMarkdown() should not contain %q\nGot:\n%s", noPart, got)
				}
			}
		})
	}
}

func TestPullRequestToMarkdown(t *testing.T) {
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	mergedTime := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		pr        *github.PullRequest
		wantParts []string
	}{
		{
			name: "basic PR",
			pr: &github.PullRequest{
				ID:        "PR_123",
				URL:       "https://github.com/owner/repo/pull/1",
				Number:    1,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Add feature",
				Body:      "This PR adds a new feature",
				State:     "open",
				Author:    "contributor",
				HeadRef:   "feature-branch",
				BaseRef:   "main",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"id: PR_123",
				"state: open",
				"head_ref: feature-branch",
				"base_ref: main",
				"# Add feature",
				"This PR adds a new feature",
				"<!-- gh-md:new-comment -->",
			},
		},
		{
			name: "merged PR",
			pr: &github.PullRequest{
				ID:          "PR_456",
				URL:         "https://github.com/owner/repo/pull/2",
				Number:      2,
				Owner:       "owner",
				Repo:        "repo",
				Title:       "Merged PR",
				Body:        "Already merged",
				State:       "merged",
				Author:      "author",
				HeadRef:     "fix-branch",
				BaseRef:     "main",
				MergeCommit: "abc123def",
				MergedAt:    mergedTime,
				CreatedAt:   baseTime,
				UpdatedAt:   mergedTime,
			},
			wantParts: []string{
				"state: merged",
				"merge_commit: abc123def",
				"merged: 2026-01-16T12:00:00Z",
			},
		},
		{
			name: "draft PR with reviewers",
			pr: &github.PullRequest{
				ID:        "PR_789",
				URL:       "https://github.com/owner/repo/pull/3",
				Number:    3,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "WIP: Draft PR",
				Body:      "Work in progress",
				State:     "open",
				Author:    "dev",
				Draft:     true,
				HeadRef:   "wip-branch",
				BaseRef:   "main",
				Reviewers: []string{"reviewer1", "reviewer2"},
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"draft: true",
				"reviewers:",
				"- reviewer1",
				"- reviewer2",
			},
		},
		{
			name: "PR with review threads",
			pr: &github.PullRequest{
				ID:        "PR_review",
				URL:       "https://github.com/owner/repo/pull/4",
				Number:    4,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "PR with Reviews",
				Body:      "Has review comments",
				State:     "open",
				Author:    "author",
				HeadRef:   "review-branch",
				BaseRef:   "main",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				ReviewThreads: []github.ReviewThread{
					{
						ID:   "PRRT_001",
						Path: "main.go",
						Line: 42,
						Comments: []github.ReviewComment{
							{
								ID:        "PRRC_001",
								Author:    "reviewer",
								Body:      "Consider refactoring this",
								CreatedAt: baseTime,
							},
						},
					},
				},
			},
			wantParts: []string{
				"## Review Threads",
				"<!-- gh-md:review-thread",
				"id: PRRT_001",
				"path: main.go",
				"line: 42",
				"### `main.go:42`",
				"<!-- gh-md:review-comment",
				"id: PRRC_001",
				"Consider refactoring this",
				"<!-- gh-md:new-comment reply_to: PRRT_001 -->",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PullRequestToMarkdown(tt.pr)
			if err != nil {
				t.Fatalf("PullRequestToMarkdown() error = %v", err)
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("PullRequestToMarkdown() missing %q\nGot:\n%s", part, got)
				}
			}
		})
	}
}

func TestDiscussionToMarkdown(t *testing.T) {
	baseTime := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		d         *github.Discussion
		wantParts []string
	}{
		{
			name: "basic discussion",
			d: &github.Discussion{
				ID:        "D_123",
				URL:       "https://github.com/owner/repo/discussions/1",
				Number:    1,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Question about feature",
				Body:      "How does this work?",
				State:     "open",
				Category:  "Q&A",
				Author:    "curious",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"id: D_123",
				"category: Q&A",
				"# Question about feature",
				"How does this work?",
				"<!-- gh-md:new-comment -->",
			},
		},
		{
			name: "discussion with answer",
			d: &github.Discussion{
				ID:        "D_456",
				URL:       "https://github.com/owner/repo/discussions/2",
				Number:    2,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Answered Question",
				Body:      "This was answered",
				State:     "open",
				Category:  "Q&A",
				Author:    "asker",
				AnswerID:  "DC_answer",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"answer_id: DC_answer",
			},
		},
		{
			name: "locked discussion",
			d: &github.Discussion{
				ID:        "D_789",
				URL:       "https://github.com/owner/repo/discussions/3",
				Number:    3,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Locked Discussion",
				Body:      "This is locked",
				State:     "open",
				Category:  "Announcements",
				Author:    "admin",
				Locked:    true,
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
			wantParts: []string{
				"locked: true",
			},
		},
		{
			name: "discussion with nested replies",
			d: &github.Discussion{
				ID:        "D_nested",
				URL:       "https://github.com/owner/repo/discussions/4",
				Number:    4,
				Owner:     "owner",
				Repo:      "repo",
				Title:     "Discussion with Replies",
				Body:      "Main topic",
				State:     "open",
				Category:  "General",
				Author:    "starter",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
				Comments: []github.DiscussionComment{
					{
						ID:        "DC_parent",
						Author:    "commenter1",
						Body:      "Top-level comment",
						CreatedAt: baseTime,
						Replies: []github.DiscussionComment{
							{
								ID:        "DC_reply",
								Author:    "commenter2",
								Body:      "Reply to comment",
								CreatedAt: baseTime,
							},
						},
					},
				},
			},
			wantParts: []string{
				"id: DC_parent",
				"Top-level comment",
				"### @commenter1",
				"<!-- gh-md:new-comment reply_to: DC_parent -->",
				"id: DC_reply",
				"parent: DC_parent",
				"Reply to comment",
				"#### @commenter2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiscussionToMarkdown(tt.d)
			if err != nil {
				t.Fatalf("DiscussionToMarkdown() error = %v", err)
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("DiscussionToMarkdown() missing %q\nGot:\n%s", part, got)
				}
			}
		})
	}
}
