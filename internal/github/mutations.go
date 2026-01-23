package github

import (
	"fmt"
	"time"
)

// Mutation queries
const (
	updateIssueMutation = `
mutation($id: ID!, $title: String!, $body: String!) {
  updateIssue(input: {id: $id, title: $title, body: $body}) {
    issue { id updatedAt }
  }
}
`

	updatePullRequestMutation = `
mutation($id: ID!, $title: String!, $body: String!) {
  updatePullRequest(input: {pullRequestId: $id, title: $title, body: $body}) {
    pullRequest { id updatedAt }
  }
}
`

	updateDiscussionMutation = `
mutation($id: ID!, $title: String!, $body: String!) {
  updateDiscussion(input: {discussionId: $id, title: $title, body: $body}) {
    discussion { id updatedAt }
  }
}
`

	closeIssueMutation = `
mutation($id: ID!) {
  closeIssue(input: {issueId: $id}) {
    issue { id state }
  }
}
`

	reopenIssueMutation = `
mutation($id: ID!) {
  reopenIssue(input: {issueId: $id}) {
    issue { id state }
  }
}
`

	closePullRequestMutation = `
mutation($id: ID!) {
  closePullRequest(input: {pullRequestId: $id}) {
    pullRequest { id state }
  }
}
`

	reopenPullRequestMutation = `
mutation($id: ID!) {
  reopenPullRequest(input: {pullRequestId: $id}) {
    pullRequest { id state }
  }
}
`

	addCommentMutation = `
mutation($subjectId: ID!, $body: String!) {
  addComment(input: {subjectId: $subjectId, body: $body}) {
    commentEdge { node { id } }
  }
}
`

	updateIssueCommentMutation = `
mutation($id: ID!, $body: String!) {
  updateIssueComment(input: {id: $id, body: $body}) {
    issueComment { id }
  }
}
`

	addDiscussionCommentMutation = `
mutation($discussionId: ID!, $body: String!) {
  addDiscussionComment(input: {discussionId: $discussionId, body: $body}) {
    comment { id }
  }
}
`

	updateDiscussionCommentMutation = `
mutation($commentId: ID!, $body: String!) {
  updateDiscussionComment(input: {commentId: $commentId, body: $body}) {
    comment { id }
  }
}
`

	addDiscussionCommentReplyMutation = `
mutation($replyToId: ID!, $body: String!) {
  addDiscussionComment(input: {replyToId: $replyToId, body: $body}) {
    comment { id }
  }
}
`

	addReviewThreadReplyMutation = `
mutation($threadId: ID!, $body: String!) {
  addPullRequestReviewThreadReply(input: {pullRequestReviewThreadId: $threadId, body: $body}) {
    comment { id }
  }
}
`

	fetchIssueCommentsQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      comments(first: 100) {
        nodes {
          id
          author { login }
          body
        }
      }
    }
  }
}
`

	fetchPullRequestCommentsQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      comments(first: 100) {
        nodes {
          id
          author { login }
          body
        }
      }
    }
  }
}
`

	fetchDiscussionCommentsQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      comments(first: 100) {
        nodes {
          id
          author { login }
          body
          replies(first: 100) {
            nodes {
              id
              author { login }
              body
            }
          }
        }
      }
    }
  }
}
`

	fetchIssueUpdatedAtQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      updatedAt
      state
    }
  }
}
`

	fetchPullRequestUpdatedAtQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      updatedAt
      state
    }
  }
}
`

	fetchDiscussionUpdatedAtQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      updatedAt
    }
  }
}
`
)

// UpdateIssue updates an issue's title and body.
func (c *Client) UpdateIssue(id, title, body string) error {
	vars := map[string]any{
		"id":    id,
		"title": title,
		"body":  body,
	}

	var resp struct {
		UpdateIssue struct {
			Issue struct {
				ID        string    `json:"id"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"issue"`
		} `json:"updateIssue"`
	}

	if err := c.Query(updateIssueMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	return nil
}

// UpdatePullRequest updates a PR's title and body.
func (c *Client) UpdatePullRequest(id, title, body string) error {
	vars := map[string]any{
		"id":    id,
		"title": title,
		"body":  body,
	}

	var resp struct {
		UpdatePullRequest struct {
			PullRequest struct {
				ID        string    `json:"id"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"pullRequest"`
		} `json:"updatePullRequest"`
	}

	if err := c.Query(updatePullRequestMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	return nil
}

// UpdateDiscussion updates a discussion's title and body.
func (c *Client) UpdateDiscussion(id, title, body string) error {
	vars := map[string]any{
		"id":    id,
		"title": title,
		"body":  body,
	}

	var resp struct {
		UpdateDiscussion struct {
			Discussion struct {
				ID        string    `json:"id"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"discussion"`
		} `json:"updateDiscussion"`
	}

	if err := c.Query(updateDiscussionMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to update discussion: %w", err)
	}

	return nil
}

// CloseIssue closes an issue.
func (c *Client) CloseIssue(id string) error {
	vars := map[string]any{
		"id": id,
	}

	var resp struct {
		CloseIssue struct {
			Issue struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"issue"`
		} `json:"closeIssue"`
	}

	if err := c.Query(closeIssueMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to close issue: %w", err)
	}

	return nil
}

// ReopenIssue reopens an issue.
func (c *Client) ReopenIssue(id string) error {
	vars := map[string]any{
		"id": id,
	}

	var resp struct {
		ReopenIssue struct {
			Issue struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"issue"`
		} `json:"reopenIssue"`
	}

	if err := c.Query(reopenIssueMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to reopen issue: %w", err)
	}

	return nil
}

// ClosePullRequest closes a pull request.
func (c *Client) ClosePullRequest(id string) error {
	vars := map[string]any{
		"id": id,
	}

	var resp struct {
		ClosePullRequest struct {
			PullRequest struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequest"`
		} `json:"closePullRequest"`
	}

	if err := c.Query(closePullRequestMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to close pull request: %w", err)
	}

	return nil
}

// ReopenPullRequest reopens a pull request.
func (c *Client) ReopenPullRequest(id string) error {
	vars := map[string]any{
		"id": id,
	}

	var resp struct {
		ReopenPullRequest struct {
			PullRequest struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"pullRequest"`
		} `json:"reopenPullRequest"`
	}

	if err := c.Query(reopenPullRequestMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to reopen pull request: %w", err)
	}

	return nil
}

// RemoteState holds the remote item's current state info.
type RemoteState struct {
	UpdatedAt time.Time
	State     string // "OPEN", "CLOSED", "MERGED" (PRs only)
}

// FetchRemoteState fetches the updatedAt timestamp and state for an item.
func (c *Client) FetchRemoteState(itemType ItemType, owner, repo string, number int) (RemoteState, error) {
	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	switch itemType {
	case ItemTypeIssue:
		var resp struct {
			Repository struct {
				Issue struct {
					UpdatedAt time.Time `json:"updatedAt"`
					State     string    `json:"state"`
				} `json:"issue"`
			} `json:"repository"`
		}
		if err := c.Query(fetchIssueUpdatedAtQuery, vars, &resp); err != nil {
			return RemoteState{}, err
		}
		return RemoteState{
			UpdatedAt: resp.Repository.Issue.UpdatedAt,
			State:     resp.Repository.Issue.State,
		}, nil

	case ItemTypePullRequest:
		var resp struct {
			Repository struct {
				PullRequest struct {
					UpdatedAt time.Time `json:"updatedAt"`
					State     string    `json:"state"`
				} `json:"pullRequest"`
			} `json:"repository"`
		}
		if err := c.Query(fetchPullRequestUpdatedAtQuery, vars, &resp); err != nil {
			return RemoteState{}, err
		}
		return RemoteState{
			UpdatedAt: resp.Repository.PullRequest.UpdatedAt,
			State:     resp.Repository.PullRequest.State,
		}, nil

	case ItemTypeDiscussion:
		var resp struct {
			Repository struct {
				Discussion struct {
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"discussion"`
			} `json:"repository"`
		}
		if err := c.Query(fetchDiscussionUpdatedAtQuery, vars, &resp); err != nil {
			return RemoteState{}, err
		}
		return RemoteState{
			UpdatedAt: resp.Repository.Discussion.UpdatedAt,
			State:     "", // Discussions don't have state
		}, nil

	default:
		return RemoteState{}, fmt.Errorf("unknown item type: %s", itemType)
	}
}

// AddComment adds a new comment to an issue or PR.
func (c *Client) AddComment(subjectID, body string) error {
	vars := map[string]any{
		"subjectId": subjectID,
		"body":      body,
	}

	var resp struct {
		AddComment struct {
			CommentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"commentEdge"`
		} `json:"addComment"`
	}

	if err := c.Query(addCommentMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

// UpdateIssueComment updates an existing comment on an issue or PR.
func (c *Client) UpdateIssueComment(commentID, body string) error {
	vars := map[string]any{
		"id":   commentID,
		"body": body,
	}

	var resp struct {
		UpdateIssueComment struct {
			IssueComment struct {
				ID string `json:"id"`
			} `json:"issueComment"`
		} `json:"updateIssueComment"`
	}

	if err := c.Query(updateIssueCommentMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// AddDiscussionComment adds a new comment to a discussion.
func (c *Client) AddDiscussionComment(discussionID, body string) error {
	vars := map[string]any{
		"discussionId": discussionID,
		"body":         body,
	}

	var resp struct {
		AddDiscussionComment struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"addDiscussionComment"`
	}

	if err := c.Query(addDiscussionCommentMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to add discussion comment: %w", err)
	}

	return nil
}

// UpdateDiscussionComment updates an existing discussion comment.
func (c *Client) UpdateDiscussionComment(commentID, body string) error {
	vars := map[string]any{
		"commentId": commentID,
		"body":      body,
	}

	var resp struct {
		UpdateDiscussionComment struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"updateDiscussionComment"`
	}

	if err := c.Query(updateDiscussionCommentMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to update discussion comment: %w", err)
	}

	return nil
}

// AddDiscussionCommentReply adds a reply to a discussion comment.
func (c *Client) AddDiscussionCommentReply(replyToID, body string) error {
	vars := map[string]any{
		"replyToId": replyToID,
		"body":      body,
	}

	var resp struct {
		AddDiscussionComment struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"addDiscussionComment"`
	}

	if err := c.Query(addDiscussionCommentReplyMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to add discussion reply: %w", err)
	}

	return nil
}

// AddReviewThreadReply adds a reply to a PR review thread.
func (c *Client) AddReviewThreadReply(threadID, body string) error {
	vars := map[string]any{
		"threadId": threadID,
		"body":     body,
	}

	var resp struct {
		AddPullRequestReviewThreadReply struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"addPullRequestReviewThreadReply"`
	}

	if err := c.Query(addReviewThreadReplyMutation, vars, &resp); err != nil {
		return fmt.Errorf("failed to add review thread reply: %w", err)
	}

	return nil
}

// RemoteComment represents a comment fetched from GitHub for comparison.
type RemoteComment struct {
	ID   string
	Body string
}

// FetchComments fetches current comments for an item.
func (c *Client) FetchComments(itemType ItemType, owner, repo string, number int) ([]RemoteComment, error) {
	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	switch itemType {
	case ItemTypeIssue:
		var resp struct {
			Repository struct {
				Issue struct {
					Comments struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body string `json:"body"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"issue"`
			} `json:"repository"`
		}
		if err := c.Query(fetchIssueCommentsQuery, vars, &resp); err != nil {
			return nil, err
		}
		var comments []RemoteComment
		for _, n := range resp.Repository.Issue.Comments.Nodes {
			comments = append(comments, RemoteComment{ID: n.ID, Body: n.Body})
		}
		return comments, nil

	case ItemTypePullRequest:
		var resp struct {
			Repository struct {
				PullRequest struct {
					Comments struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body string `json:"body"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"pullRequest"`
			} `json:"repository"`
		}
		if err := c.Query(fetchPullRequestCommentsQuery, vars, &resp); err != nil {
			return nil, err
		}
		var comments []RemoteComment
		for _, n := range resp.Repository.PullRequest.Comments.Nodes {
			comments = append(comments, RemoteComment{ID: n.ID, Body: n.Body})
		}
		return comments, nil

	case ItemTypeDiscussion:
		var resp struct {
			Repository struct {
				Discussion struct {
					Comments struct {
						Nodes []struct {
							ID     string `json:"id"`
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body    string `json:"body"`
							Replies struct {
								Nodes []struct {
									ID     string `json:"id"`
									Author struct {
										Login string `json:"login"`
									} `json:"author"`
									Body string `json:"body"`
								} `json:"nodes"`
							} `json:"replies"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"discussion"`
			} `json:"repository"`
		}
		if err := c.Query(fetchDiscussionCommentsQuery, vars, &resp); err != nil {
			return nil, err
		}
		var comments []RemoteComment
		for _, n := range resp.Repository.Discussion.Comments.Nodes {
			comments = append(comments, RemoteComment{ID: n.ID, Body: n.Body})
			for _, r := range n.Replies.Nodes {
				comments = append(comments, RemoteComment{ID: r.ID, Body: r.Body})
			}
		}
		return comments, nil

	default:
		return nil, fmt.Errorf("unknown item type: %s", itemType)
	}
}
