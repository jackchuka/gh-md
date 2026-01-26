package github

import (
	"time"
)

// Discussions query uses smaller limits due to GitHub's 500k node limit
// (discussions × comments × replies must stay under 500k)
const discussionsQuery = `
query($owner: String!, $repo: String!, $first: Int!, $after: String, $states: [DiscussionState!]) {
  repository(owner: $owner, name: $repo) {
    discussions(first: $first, after: $after, states: $states, orderBy: {field: UPDATED_AT, direction: DESC}) {
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
        closed
        locked
        createdAt
        updatedAt
        category {
          name
        }
        author {
          login
        }
        answer {
          id
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
            replies(first: 20) {
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

const singleDiscussionQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      id
      url
      number
      title
      body
      closed
      locked
      createdAt
      updatedAt
      category {
        name
      }
      author {
        login
      }
      answer {
        id
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
          replies(first: 100) {
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

// DiscussionsResponse represents the GraphQL response for discussions.
type DiscussionsResponse struct {
	Repository struct {
		Discussions struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []DiscussionNode `json:"nodes"`
		} `json:"discussions"`
	} `json:"repository"`
}

// SingleDiscussionResponse represents the GraphQL response for a single discussion.
type SingleDiscussionResponse struct {
	Repository struct {
		Discussion DiscussionNode `json:"discussion"`
	} `json:"repository"`
}

// DiscussionNode represents a discussion in the GraphQL response.
type DiscussionNode struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Closed    bool      `json:"closed"`
	Locked    bool      `json:"locked"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Category  struct {
		Name string `json:"name"`
	} `json:"category"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Answer struct {
		ID string `json:"id"`
	} `json:"answer"`
	Comments struct {
		Nodes []DiscussionCommentNode `json:"nodes"`
	} `json:"comments"`
}

// DiscussionCommentNode represents a comment in a discussion.
type DiscussionCommentNode struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Replies struct {
		Nodes []struct {
			ID        string    `json:"id"`
			Body      string    `json:"body"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"replies"`
}

// FetchDiscussion fetches a single discussion by number.
func (c *Client) FetchDiscussion(owner, repo string, number int) (*Discussion, error) {
	vars := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var resp SingleDiscussionResponse
	if err := c.Query(singleDiscussionQuery, vars, &resp); err != nil {
		return nil, err
	}

	return nodeToDiscussion(resp.Repository.Discussion, owner, repo), nil
}

// FetchDiscussions fetches all discussions from a repository with pagination.
// If openOnly is true, only OPEN discussions are fetched; otherwise all states are fetched.
// If since is provided, fetching stops when encountering items older than the timestamp.
func (c *Client) FetchDiscussions(owner, repo string, limit int, openOnly bool, since *time.Time) ([]Discussion, error) {
	var discussions []Discussion
	var cursor *string
	// Smaller page size for discussions due to nested comments/replies
	pageSize := 10
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	var states []string
	if openOnly {
		states = []string{"OPEN"}
	} else {
		states = []string{"OPEN", "CLOSED"}
	}

	for {
		vars := map[string]any{
			"owner":  owner,
			"repo":   repo,
			"first":  pageSize,
			"states": states,
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		var resp DiscussionsResponse
		if err := c.Query(discussionsQuery, vars, &resp); err != nil {
			return nil, err
		}

		for _, node := range resp.Repository.Discussions.Nodes {
			// Early termination: stop if item is older than since timestamp
			if since != nil && node.UpdatedAt.Before(*since) {
				return discussions, nil
			}
			discussions = append(discussions, *nodeToDiscussion(node, owner, repo))
			if limit > 0 && len(discussions) >= limit {
				return discussions, nil
			}
		}

		if !resp.Repository.Discussions.PageInfo.HasNextPage {
			break
		}
		cursor = &resp.Repository.Discussions.PageInfo.EndCursor
	}

	return discussions, nil
}

func nodeToDiscussion(node DiscussionNode, owner, repo string) *Discussion {
	comments := make([]DiscussionComment, 0, len(node.Comments.Nodes))
	for _, c := range node.Comments.Nodes {
		replies := make([]DiscussionComment, 0, len(c.Replies.Nodes))
		for _, r := range c.Replies.Nodes {
			replies = append(replies, DiscussionComment{
				ID:        r.ID,
				Author:    r.Author.Login,
				Body:      r.Body,
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			})
		}

		comments = append(comments, DiscussionComment{
			ID:        c.ID,
			Author:    c.Author.Login,
			Body:      c.Body,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Replies:   replies,
		})
	}

	state := "open"
	if node.Closed {
		state = "closed"
	}

	return &Discussion{
		ID:        node.ID,
		URL:       node.URL,
		Number:    node.Number,
		Owner:     owner,
		Repo:      repo,
		Title:     node.Title,
		Body:      node.Body,
		State:     state,
		Category:  node.Category.Name,
		Author:    node.Author.Login,
		AnswerID:  node.Answer.ID,
		Locked:    node.Locked,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
		Comments:  comments,
	}
}
