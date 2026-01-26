package github

import "time"

// Comment represents a comment on an issue, PR, or discussion.
type Comment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ReviewThread represents a review thread on a PR (a conversation on a specific line).
type ReviewThread struct {
	ID         string          `json:"id"`
	Path       string          `json:"path"`
	Line       int             `json:"line"`
	IsResolved bool            `json:"isResolved"`
	IsOutdated bool            `json:"isOutdated"`
	Comments   []ReviewComment `json:"comments"`
}

// ReviewComment represents an inline review comment on a PR.
type ReviewComment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// IssueReference represents a reference to a parent or child issue.
type IssueReference struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	State  string `json:"state"`
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
}

// SubIssuesSummary provides statistics about sub-issues.
type SubIssuesSummary struct {
	Total           int `json:"total"`
	Completed       int `json:"completed"`
	PercentComplete int `json:"percentComplete"`
}

// Issue represents a GitHub issue with all metadata.
type Issue struct {
	ID               string            `json:"id"`
	URL              string            `json:"url"`
	Number           int               `json:"number"`
	Owner            string            `json:"owner"`
	Repo             string            `json:"repo"`
	Title            string            `json:"title"`
	Body             string            `json:"body"`
	State            string            `json:"state"`
	Author           string            `json:"author"`
	Labels           []string          `json:"labels"`
	Assignees        []string          `json:"assignees"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	Comments         []Comment         `json:"comments"`
	Parent           *IssueReference   `json:"parent,omitempty"`
	Children         []IssueReference  `json:"children,omitempty"`
	SubIssuesSummary *SubIssuesSummary `json:"subIssuesSummary,omitempty"`
}

// PullRequest represents a GitHub pull request with all metadata.
type PullRequest struct {
	ID            string         `json:"id"`
	URL           string         `json:"url"`
	Number        int            `json:"number"`
	Owner         string         `json:"owner"`
	Repo          string         `json:"repo"`
	Title         string         `json:"title"`
	Body          string         `json:"body"`
	State         string         `json:"state"`
	Author        string         `json:"author"`
	Draft         bool           `json:"draft"`
	Labels        []string       `json:"labels"`
	Assignees     []string       `json:"assignees"`
	HeadRef       string         `json:"headRef"`
	BaseRef       string         `json:"baseRef"`
	MergeCommit   string         `json:"mergeCommit,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	MergedAt      time.Time      `json:"mergedAt,omitempty"`
	Comments      []Comment      `json:"comments"`
	ReviewThreads []ReviewThread `json:"reviewThreads"`
}

// DiscussionComment represents a comment or reply in a discussion.
type DiscussionComment struct {
	ID        string              `json:"id"`
	Author    string              `json:"author"`
	Body      string              `json:"body"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
	Replies   []DiscussionComment `json:"replies,omitempty"`
}

// Discussion represents a GitHub discussion with all metadata.
type Discussion struct {
	ID        string              `json:"id"`
	URL       string              `json:"url"`
	Number    int                 `json:"number"`
	Owner     string              `json:"owner"`
	Repo      string              `json:"repo"`
	Title     string              `json:"title"`
	Body      string              `json:"body"`
	State     string              `json:"state"`
	Category  string              `json:"category"`
	Author    string              `json:"author"`
	AnswerID  string              `json:"answerId,omitempty"`
	Locked    bool                `json:"locked"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
	Comments  []DiscussionComment `json:"comments"`
}

// ItemType represents the type of GitHub item.
type ItemType string

const (
	ItemTypeIssue       ItemType = "issue"
	ItemTypePullRequest ItemType = "pull"
	ItemTypeDiscussion  ItemType = "discussion"
)

// ParsedInput represents parsed command input (URL or owner/repo).
type ParsedInput struct {
	Owner    string
	Repo     string
	Number   int      // 0 if fetching all
	ItemType ItemType // Empty if fetching all types
}
