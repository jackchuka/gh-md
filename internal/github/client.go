package github

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Client provides methods to interact with GitHub's GraphQL API.
type Client struct {
	gql *api.GraphQLClient
}

// NewClient creates a new GitHub client using gh auth.
func NewClient() (*Client, error) {
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	return &Client{gql: gql}, nil
}

// URL patterns for GitHub resources.
var (
	// Matches: https://github.com/owner/repo/issues/123
	issueURLPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/issues/(\d+)/?$`)
	// Matches: https://github.com/owner/repo/pull/123
	pullURLPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/pull/(\d+)/?$`)
	// Matches: https://github.com/owner/repo/discussions/123
	discussionURLPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/discussions/(\d+)/?$`)
	// Matches: owner/repo
	ownerRepoPattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)
)

// ParseInput parses the input argument which can be a URL or owner/repo format.
func ParseInput(input string) (*ParsedInput, error) {
	input = strings.TrimSpace(input)

	// Try issue URL
	if matches := issueURLPattern.FindStringSubmatch(input); matches != nil {
		number, _ := strconv.Atoi(matches[3])
		return &ParsedInput{
			Owner:    matches[1],
			Repo:     matches[2],
			Number:   number,
			ItemType: ItemTypeIssue,
		}, nil
	}

	// Try PR URL
	if matches := pullURLPattern.FindStringSubmatch(input); matches != nil {
		number, _ := strconv.Atoi(matches[3])
		return &ParsedInput{
			Owner:    matches[1],
			Repo:     matches[2],
			Number:   number,
			ItemType: ItemTypePullRequest,
		}, nil
	}

	// Try discussion URL
	if matches := discussionURLPattern.FindStringSubmatch(input); matches != nil {
		number, _ := strconv.Atoi(matches[3])
		return &ParsedInput{
			Owner:    matches[1],
			Repo:     matches[2],
			Number:   number,
			ItemType: ItemTypeDiscussion,
		}, nil
	}

	// Try owner/repo format
	if matches := ownerRepoPattern.FindStringSubmatch(input); matches != nil {
		return &ParsedInput{
			Owner: matches[1],
			Repo:  matches[2],
		}, nil
	}

	return nil, fmt.Errorf("invalid input format: %s (expected URL or owner/repo)", input)
}

// Query executes a GraphQL query.
func (c *Client) Query(query string, variables map[string]interface{}, response interface{}) error {
	return c.gql.Do(query, variables, response)
}
