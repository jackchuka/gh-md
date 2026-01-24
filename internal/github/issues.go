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
        parent {
          id
          number
          title
          url
          state
          repository {
            owner { login }
            name
          }
        }
        subIssues(first: 50) {
          nodes {
            id
            number
            title
            url
            state
            repository {
              owner { login }
              name
            }
          }
        }
        subIssuesSummary {
          total
          completed
          percentCompleted
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
      parent {
        id
        number
        title
        url
        state
        repository {
          owner { login }
          name
        }
      }
      subIssues(first: 50) {
        nodes {
          id
          number
          title
          url
          state
          repository {
            owner { login }
            name
          }
        }
      }
      subIssuesSummary {
        total
        completed
        percentCompleted
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

// ParentIssueNode represents a parent issue in the GraphQL response.
type ParentIssueNode struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	State      string `json:"state"`
	Repository struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	} `json:"repository"`
}

// SubIssueNode represents a sub-issue in the GraphQL response.
type SubIssueNode struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	State      string `json:"state"`
	Repository struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	} `json:"repository"`
}

// SubIssuesSummaryNode represents the sub-issues summary in the GraphQL response.
type SubIssuesSummaryNode struct {
	Total            int `json:"total"`
	Completed        int `json:"completed"`
	PercentCompleted int `json:"percentCompleted"`
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
	Parent    *ParentIssueNode `json:"parent"`
	SubIssues struct {
		Nodes []SubIssueNode `json:"nodes"`
	} `json:"subIssues"`
	SubIssuesSummary *SubIssuesSummaryNode `json:"subIssuesSummary"`
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

	issue := &Issue{
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

	// Convert parent issue
	if node.Parent != nil {
		issue.Parent = &IssueReference{
			ID:     node.Parent.ID,
			Number: node.Parent.Number,
			Title:  node.Parent.Title,
			URL:    node.Parent.URL,
			State:  strings.ToLower(node.Parent.State),
			Owner:  node.Parent.Repository.Owner.Login,
			Repo:   node.Parent.Repository.Name,
		}
	}

	// Convert sub-issues (children)
	if len(node.SubIssues.Nodes) > 0 {
		issue.Children = make([]IssueReference, 0, len(node.SubIssues.Nodes))
		for _, sub := range node.SubIssues.Nodes {
			issue.Children = append(issue.Children, IssueReference{
				ID:     sub.ID,
				Number: sub.Number,
				Title:  sub.Title,
				URL:    sub.URL,
				State:  strings.ToLower(sub.State),
				Owner:  sub.Repository.Owner.Login,
				Repo:   sub.Repository.Name,
			})
		}
	}

	// Convert sub-issues summary
	if node.SubIssuesSummary != nil && node.SubIssuesSummary.Total > 0 {
		issue.SubIssuesSummary = &SubIssuesSummary{
			Total:           node.SubIssuesSummary.Total,
			Completed:       node.SubIssuesSummary.Completed,
			PercentComplete: node.SubIssuesSummary.PercentCompleted,
		}
	}

	return issue
}
