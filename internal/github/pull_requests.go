package github

import (
	"strings"
	"time"
)

// pullRequestsQuery fetches multiple PRs with pagination.
// GitHub GraphQL API has a 500,000 node limit per query.
// With nested connections, we must limit: 25 PRs × (50 comments + 50 threads × 20 nested) = 26,250 nodes
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
        author {
          login
        }
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
        comments(first: 50) {
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
        reviewThreads(first: 50) {
          nodes {
            id
            path
            line
            isResolved
            isOutdated
            comments(first: 20) {
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
      author {
        login
      }
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
      reviewThreads(first: 100) {
        nodes {
          id
          path
          line
          isResolved
          isOutdated
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
	ID     string `json:"id"`
	URL    string `json:"url"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
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
	ReviewThreads struct {
		Nodes []struct {
			ID         string `json:"id"`
			Path       string `json:"path"`
			Line       int    `json:"line"`
			IsResolved bool   `json:"isResolved"`
			IsOutdated bool   `json:"isOutdated"`
			Comments   struct {
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
		} `json:"nodes"`
	} `json:"reviewThreads"`
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
	pageSize := 25
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

	// Extract review threads
	var reviewThreads []ReviewThread
	for _, thread := range node.ReviewThreads.Nodes {
		var threadComments []ReviewComment
		for _, c := range thread.Comments.Nodes {
			threadComments = append(threadComments, ReviewComment{
				ID:        c.ID,
				Author:    c.Author.Login,
				Body:      c.Body,
				CreatedAt: c.CreatedAt,
				UpdatedAt: c.UpdatedAt,
			})
		}
		reviewThreads = append(reviewThreads, ReviewThread{
			ID:         thread.ID,
			Path:       thread.Path,
			Line:       thread.Line,
			IsResolved: thread.IsResolved,
			IsOutdated: thread.IsOutdated,
			Comments:   threadComments,
		})
	}

	return &PullRequest{
		ID:            node.ID,
		URL:           node.URL,
		Number:        node.Number,
		Owner:         owner,
		Repo:          repo,
		Title:         node.Title,
		Body:          node.Body,
		State:         strings.ToLower(node.State),
		Author:        node.Author.Login,
		Draft:         node.IsDraft,
		Labels:        labels,
		Assignees:     assignees,
		HeadRef:       node.HeadRefName,
		BaseRef:       node.BaseRefName,
		MergeCommit:   node.MergeCommit.Oid,
		CreatedAt:     node.CreatedAt,
		UpdatedAt:     node.UpdatedAt,
		MergedAt:      node.MergedAt,
		Comments:      comments,
		ReviewThreads: reviewThreads,
	}
}
