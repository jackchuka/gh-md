package github

import (
	"strings"
	"time"
)

const pullRequestsQuery = `
query($owner: String!, $repo: String!, $first: Int!, $after: String, $states: [PullRequestState!]) {
  repository(owner: $owner, name: $repo) {
    pullRequests(first: $first, after: $after, states: $states, orderBy: {field: UPDATED_AT, direction: DESC}) {
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        id
        url
        number
        title
        body
        state
        isDraft
        createdAt
        updatedAt
        mergedAt
        headRefName
        baseRefName
        mergeCommit {
          oid
        }
        labels(first: 100) {
          nodes {
            name
          }
        }
        assignees(first: 100) {
          nodes {
            login
          }
        }
        comments(first: 100) {
          nodes {
            id
            body
            createdAt
            updatedAt
            author {
              login
            }
          }
        }
      }
    }
  }
}
`

const singlePullRequestQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      id
      url
      number
      title
      body
      state
      isDraft
      createdAt
      updatedAt
      mergedAt
      headRefName
      baseRefName
      mergeCommit {
        oid
      }
      labels(first: 100) {
        nodes {
          name
        }
      }
      assignees(first: 100) {
        nodes {
          login
        }
      }
      comments(first: 100) {
        nodes {
          id
          body
          createdAt
          updatedAt
          author {
            login
          }
        }
      }
    }
  }
}
`

// PullRequestsResponse represents the GraphQL response for pull requests.
type PullRequestsResponse struct {
	Repository struct {
		PullRequests struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []PullRequestNode `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

// SinglePullRequestResponse represents the GraphQL response for a single PR.
type SinglePullRequestResponse struct {
	Repository struct {
		PullRequest PullRequestNode `json:"pullRequest"`
	} `json:"repository"`
}

// PullRequestNode represents a PR in the GraphQL response.
type PullRequestNode struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	IsDraft     bool      `json:"isDraft"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	MergedAt    time.Time `json:"mergedAt"`
	HeadRefName string    `json:"headRefName"`
	BaseRefName string    `json:"baseRefName"`
	MergeCommit struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
	Comments struct {
		Nodes []struct {
			ID        string    `json:"id"`
			Body      string    `json:"body"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"comments"`
}

// FetchPullRequest fetches a single PR by number.
func (c *Client) FetchPullRequest(owner, repo string, number int) (*PullRequest, error) {
	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var resp SinglePullRequestResponse
	if err := c.Query(singlePullRequestQuery, vars, &resp); err != nil {
		return nil, err
	}

	return nodeToPullRequest(resp.Repository.PullRequest, owner, repo), nil
}

// FetchPullRequests fetches all PRs from a repository with pagination.
func (c *Client) FetchPullRequests(owner, repo string, limit int) ([]PullRequest, error) {
	var prs []PullRequest
	var cursor *string
	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	for {
		vars := map[string]any{
			"owner":  owner,
			"repo":   repo,
			"first":  pageSize,
			"states": []string{"OPEN", "CLOSED", "MERGED"},
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		var resp PullRequestsResponse
		if err := c.Query(pullRequestsQuery, vars, &resp); err != nil {
			return nil, err
		}

		for _, node := range resp.Repository.PullRequests.Nodes {
			prs = append(prs, *nodeToPullRequest(node, owner, repo))
			if limit > 0 && len(prs) >= limit {
				return prs, nil
			}
		}

		if !resp.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		cursor = &resp.Repository.PullRequests.PageInfo.EndCursor
	}

	return prs, nil
}

func nodeToPullRequest(node PullRequestNode, owner, repo string) *PullRequest {
	labels := make([]string, 0, len(node.Labels.Nodes))
	for _, l := range node.Labels.Nodes {
		labels = append(labels, l.Name)
	}

	assignees := make([]string, 0, len(node.Assignees.Nodes))
	for _, a := range node.Assignees.Nodes {
		assignees = append(assignees, a.Login)
	}

	comments := make([]Comment, 0, len(node.Comments.Nodes))
	for _, c := range node.Comments.Nodes {
		comments = append(comments, Comment{
			ID:        c.ID,
			Author:    c.Author.Login,
			Body:      c.Body,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	return &PullRequest{
		ID:          node.ID,
		URL:         node.URL,
		Number:      node.Number,
		Owner:       owner,
		Repo:        repo,
		Title:       node.Title,
		Body:        node.Body,
		State:       strings.ToLower(node.State),
		Draft:       node.IsDraft,
		Labels:      labels,
		Assignees:   assignees,
		HeadRef:     node.HeadRefName,
		BaseRef:     node.BaseRefName,
		MergeCommit: node.MergeCommit.Oid,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
		MergedAt:    node.MergedAt,
		Comments:    comments,
	}
}
