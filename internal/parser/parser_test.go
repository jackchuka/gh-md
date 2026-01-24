package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseComments_ExistingMarkerComment(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: IC_abc123
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Existing comment body
<!-- /gh-md:comment -->

---
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(parsed.Comments))
	}

	c := parsed.Comments[0]
	if c.ID != "IC_abc123" {
		t.Errorf("expected ID 'IC_abc123', got %q", c.ID)
	}
	if c.Body != "Existing comment body" {
		t.Errorf("expected body 'Existing comment body', got %q", c.Body)
	}
	if c.ParentID != "" {
		t.Errorf("expected no ParentID, got %q", c.ParentID)
	}
}

func TestParseComments_NewCommentMarker(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: IC_abc123
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Existing comment
<!-- /gh-md:comment -->

<!-- gh-md:new-comment -->
New comment using marker format
<!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(parsed.Comments))
	}

	// First is existing
	if parsed.Comments[0].ID != "IC_abc123" {
		t.Errorf("expected first comment ID 'IC_abc123', got %q", parsed.Comments[0].ID)
	}

	// Second is new (no ID)
	newComment := parsed.Comments[1]
	if newComment.ID != "" {
		t.Errorf("expected new comment to have no ID, got %q", newComment.ID)
	}
	if newComment.Body != "New comment using marker format" {
		t.Errorf("expected body 'New comment using marker format', got %q", newComment.Body)
	}
}

func TestParseComments_EmptyNewCommentMarkerIgnored(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: IC_abc123
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Existing comment
<!-- /gh-md:comment -->

<!-- gh-md:new-comment -->

<!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	// Should only have 1 comment (the existing one), empty marker ignored
	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment (empty marker ignored), got %d", len(parsed.Comments))
	}
}

func TestParseComments_MultipleNewCommentMarkers(t *testing.T) {
	// Multiple new comment markers are all parsed
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

<!-- gh-md:new-comment -->
First new comment
<!-- /gh-md:new-comment -->

<!-- gh-md:new-comment -->
Second new comment
<!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(parsed.Comments))
	}

	if parsed.Comments[0].Body != "First new comment" {
		t.Errorf("expected body 'First new comment', got %q", parsed.Comments[0].Body)
	}
	if parsed.Comments[1].Body != "Second new comment" {
		t.Errorf("expected body 'Second new comment', got %q", parsed.Comments[1].Body)
	}
}

func TestParseComments_ReplyWithReplyTo(t *testing.T) {
	// For replies, use reply_to attribute
	content := `---
id: D_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: DC_parent
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Parent comment
<!-- /gh-md:comment -->

  <!-- gh-md:new-comment reply_to: DC_parent -->
  This is a reply
  <!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "discussions/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(parsed.Comments))
	}

	reply := parsed.Comments[1]
	if reply.ParentID != "DC_parent" {
		t.Errorf("expected reply ParentID 'DC_parent', got %q", reply.ParentID)
	}
	if reply.Body != "This is a reply" {
		t.Errorf("expected reply body 'This is a reply', got %q", reply.Body)
	}
}

func TestParseComments_NoNewCommentMarker(t *testing.T) {
	// File with no new comment markers - only existing comments
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body with some text
<!-- /gh-md:content -->

---

<!-- gh-md:comment
id: IC_abc123
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Existing comment
<!-- /gh-md:comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	// Should have 1 comment (the existing one)
	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(parsed.Comments))
	}
	if parsed.Comments[0].ID != "IC_abc123" {
		t.Errorf("expected ID 'IC_abc123', got %q", parsed.Comments[0].ID)
	}
}

func TestParseComments_NewCommentNoExisting(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

<!-- gh-md:new-comment -->
New comment on item with no existing comments
<!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(parsed.Comments))
	}

	if parsed.Comments[0].Body != "New comment on item with no existing comments" {
		t.Errorf("unexpected body: %q", parsed.Comments[0].Body)
	}
}

func TestParseState(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "open state",
			content: `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---
`,
			expected: "open",
		},
		{
			name: "closed state",
			content: `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: closed
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---
`,
			expected: "closed",
		},
		{
			name: "OPEN state (uppercase)",
			content: `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: OPEN
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---
`,
			expected: "OPEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseContent(tt.content, "issues/1.md")
			if err != nil {
				t.Fatalf("parseContent failed: %v", err)
			}

			if parsed.State != tt.expected {
				t.Errorf("expected state %q, got %q", tt.expected, parsed.State)
			}
		})
	}
}

func TestParseTitleAndBody(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# My Issue Title

This is the body.
With multiple lines.
<!-- /gh-md:content -->

---
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if parsed.Title != "My Issue Title" {
		t.Errorf("expected title 'My Issue Title', got %q", parsed.Title)
	}

	expectedBody := "This is the body.\nWith multiple lines."
	if parsed.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, parsed.Body)
	}
}

func TestParseComments_CommentsMarkerNotConfusedWithComment(t *testing.T) {
	// Ensure <!-- gh-md:comments --> (plural) is not confused with <!-- gh-md:comment -->
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

---

<!-- gh-md:comments -->

<!-- gh-md:comment
id: IC_abc123
author: user1
created: 2026-01-01T00:00:00Z
-->
### @user1 (2026-01-01)

Comment body
<!-- /gh-md:comment -->

<!-- /gh-md:comments -->

---
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	// Should only find 1 comment (the actual comment, not the wrapper)
	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(parsed.Comments))
	}

	if parsed.Comments[0].ID != "IC_abc123" {
		t.Errorf("expected ID 'IC_abc123', got %q", parsed.Comments[0].ID)
	}
}

func TestDetectItemType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"owner/repo/issues/1.md", "issue"},
		{"owner/repo/pulls/1.md", "pull"},
		{"owner/repo/discussions/1.md", "discussion"},
		{"some/other/path/1.md", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectItemType(tt.path)
			if string(result) != tt.expected {
				t.Errorf("detectItemType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestParseComments_ReviewThreadReply(t *testing.T) {
	content := `---
id: PR_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

<!-- gh-md:review-thread
id: PRRT_thread123
path: file.go
line: 10
-->
### ` + "`file.go:10`" + `

<!-- gh-md:review-comment
id: PRRC_comment123
author: user1
created: 2026-01-01T00:00:00Z
-->
#### @user1 (2026-01-01)

existing review comment
<!-- /gh-md:review-comment -->

<!-- gh-md:new-comment reply_to: PRRT_thread123 -->
my reply to the review thread
<!-- /gh-md:new-comment -->

<!-- /gh-md:review-thread -->
`
	parsed, err := parseContent(content, "pulls/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	// Should have 1 new comment (the reply) - review comments are read-only
	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment (the new reply), got %d", len(parsed.Comments))
	}

	reply := parsed.Comments[0]
	if reply.ID != "" {
		t.Errorf("expected new comment to have no ID, got %q", reply.ID)
	}

	if reply.ParentID != "PRRT_thread123" {
		t.Errorf("expected ParentID 'PRRT_thread123', got %q", reply.ParentID)
	}

	if reply.Body != "my reply to the review thread" {
		t.Errorf("expected body 'my reply to the review thread', got %q", reply.Body)
	}
}

func TestResolveFilePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GH_MD_ROOT", root)

	expected := filepath.Join(root, "owner", "repo", "issues", "123.md")
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(expected, []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	t.Run("root-relative path with .md", func(t *testing.T) {
		got, err := ResolveFilePath("owner/repo/issues/123.md")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	})

	t.Run("root-relative path with .md and dot-slash", func(t *testing.T) {
		got, err := ResolveFilePath("./owner/repo/issues/123.md")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	})

	t.Run("root-relative path without .md", func(t *testing.T) {
		got, err := ResolveFilePath("owner/repo/issues/123")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	})

	t.Run("short path", func(t *testing.T) {
		got, err := ResolveFilePath("owner/repo/issues/123")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	})

	t.Run("url", func(t *testing.T) {
		got, err := ResolveFilePath("https://github.com/owner/repo/issues/123")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != expected {
			t.Fatalf("got %q, want %q", got, expected)
		}
	})

	t.Run("pull path with .md", func(t *testing.T) {
		pr := filepath.Join(root, "owner", "repo", "pulls", "456.md")
		if err := os.MkdirAll(filepath.Dir(pr), 0o755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		if err := os.WriteFile(pr, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		got, err := ResolveFilePath("owner/repo/pull/456.md")
		if err != nil {
			t.Fatalf("ResolveFilePath failed: %v", err)
		}
		if got != pr {
			t.Fatalf("got %q, want %q", got, pr)
		}
	})
}

func TestParseComments_MultilineBody(t *testing.T) {
	content := `---
id: I_123
owner: test
repo: demo
number: 1
updated: 2026-01-01T00:00:00Z
state: open
---

<!-- gh-md:content -->
# Title
Body
<!-- /gh-md:content -->

<!-- gh-md:new-comment -->
This is a multi-line comment.

It has paragraphs.

And more content.
<!-- /gh-md:new-comment -->
`
	parsed, err := parseContent(content, "issues/1.md")
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(parsed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(parsed.Comments))
	}

	expectedBody := "This is a multi-line comment.\n\nIt has paragraphs.\n\nAnd more content."
	if parsed.Comments[0].Body != expectedBody {
		t.Errorf("expected body:\n%q\ngot:\n%q", expectedBody, parsed.Comments[0].Body)
	}
}
