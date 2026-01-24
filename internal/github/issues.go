package github

import (
	"strings"
	"time"
)

const issuesQuery = `
query($owner: String!, $repo: String!, $first: Int!, $after: String, $states: [IssueState!]) {
  repository(owner: $owner, name: $repo) {
    issues(first: $first, after: $after, states: $states, orderBy: {field: UPDATED_AT, direction: DESC}) {
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
        createdAt
        updatedAt
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

const singleIssueQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
      url
      number
      title
      body
      state
      author {
        login
      }
      createdAt
      updatedAt
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

// IssuesResponse represents the GraphQL response for issues.
type IssuesResponse struct {
	Repository struct {
		Issues struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []IssueNode `json:"nodes"`
		} `json:"issues"`
	} `json:"repository"`
}

// SingleIssueResponse represents the GraphQL response for a single issue.
type SingleIssueResponse struct {
	Repository struct {
		Issue IssueNode `json:"issue"`
	} `json:"repository"`
}

// IssueNode represents an issue in the GraphQL response.
type IssueNode struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Labels    struct {
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

// FetchIssue fetches a single issue by number.
func (c *Client) FetchIssue(owner, repo string, number int) (*Issue, error) {
	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var resp SingleIssueResponse
	if err := c.Query(singleIssueQuery, vars, &resp); err != nil {
		return nil, err
	}

	return nodeToIssue(resp.Repository.Issue, owner, repo), nil
}

// FetchIssues fetches all issues from a repository with pagination.
func (c *Client) FetchIssues(owner, repo string, limit int) ([]Issue, error) {
	var issues []Issue
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
			"states": []string{"OPEN", "CLOSED"},
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		var resp IssuesResponse
		if err := c.Query(issuesQuery, vars, &resp); err != nil {
			return nil, err
		}

		for _, node := range resp.Repository.Issues.Nodes {
			issues = append(issues, *nodeToIssue(node, owner, repo))
			if limit > 0 && len(issues) >= limit {
				return issues, nil
			}
		}

		if !resp.Repository.Issues.PageInfo.HasNextPage {
			break
		}
		cursor = &resp.Repository.Issues.PageInfo.EndCursor
	}

	return issues, nil
}

func nodeToIssue(node IssueNode, owner, repo string) *Issue {
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

	return &Issue{
		ID:        node.ID,
		URL:       node.URL,
		Number:    node.Number,
		Owner:     owner,
		Repo:      repo,
		Title:     node.Title,
		Body:      node.Body,
		State:     strings.ToLower(node.State),
		Author:    node.Author.Login,
		Labels:    labels,
		Assignees: assignees,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
		Comments:  comments,
	}
}
