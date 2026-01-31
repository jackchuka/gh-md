package github

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jackchuka/gh-md/internal/discovery"
)

// Client provides methods to interact with GitHub's GraphQL API.
type Client struct {
	gql *api.GraphQLClient
}

// NewClient creates a new GitHub client using gh auth.
func NewClient() (*Client, error) {
	opts := api.ClientOptions{
		Headers: map[string]string{
			"GraphQL-Features": "sub_issues",
		},
	}
	gql, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	return &Client{gql: gql}, nil
}

// URL patterns for GitHub resources.
var (
	// Matches: https://github.com/owner/repo/issues/123 or owner/repo/issues/123 (optional .md suffix)
	issueURLPattern = regexp.MustCompile(`^(?:https?://github\.com/)?([^/]+)/([^/]+)/issues/(\d+)(?:\.md)?/?$`)
	// Matches: https://github.com/owner/repo/pull/123 or owner/repo/pull/123 (optional .md suffix)
	// Also accepts "pulls" to match local storage layout (owner/repo/pulls/123.md).
	pullURLPattern = regexp.MustCompile(`^(?:https?://github\.com/)?([^/]+)/([^/]+)/(?:pull|pulls)/(\d+)(?:\.md)?/?$`)
	// Matches: https://github.com/owner/repo/discussions/123 or owner/repo/discussions/123 (optional .md suffix)
	discussionURLPattern = regexp.MustCompile(`^(?:https?://github\.com/)?([^/]+)/([^/]+)/discussions/(\d+)(?:\.md)?/?$`)
	// Matches: owner/repo
	ownerRepoPattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)
)

// ParseInput parses the input argument which can be a URL or owner/repo format.
func ParseInput(input string) (*ParsedInput, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("invalid input format: %s (expected URL, owner/repo, or owner/repo/<type>/<number>)", input)
	}

	normalized := strings.TrimPrefix(input, "./")
	candidates := []string{input}
	if normalized != input {
		candidates = append(candidates, normalized)
	}

	for _, candidate := range candidates {
		// Try issue URL / short path
		if matches := issueURLPattern.FindStringSubmatch(candidate); matches != nil {
			number, _ := strconv.Atoi(matches[3])
			return &ParsedInput{
				Owner:    matches[1],
				Repo:     matches[2],
				Number:   number,
				ItemType: ItemTypeIssue,
			}, nil
		}

		// Try PR URL / short path
		if matches := pullURLPattern.FindStringSubmatch(candidate); matches != nil {
			number, _ := strconv.Atoi(matches[3])
			return &ParsedInput{
				Owner:    matches[1],
				Repo:     matches[2],
				Number:   number,
				ItemType: ItemTypePullRequest,
			}, nil
		}

		// Try discussion URL / short path
		if matches := discussionURLPattern.FindStringSubmatch(candidate); matches != nil {
			number, _ := strconv.Atoi(matches[3])
			return &ParsedInput{
				Owner:    matches[1],
				Repo:     matches[2],
				Number:   number,
				ItemType: ItemTypeDiscussion,
			}, nil
		}

		// Try owner/repo format
		if matches := ownerRepoPattern.FindStringSubmatch(candidate); matches != nil {
			return &ParsedInput{
				Owner: matches[1],
				Repo:  matches[2],
			}, nil
		}

		// Try path-like formats (e.g. ./owner/repo/issues/123.md, /.../owner/repo/issues/123.md).
		if parsed, ok := parsePathLikeInput(candidate); ok {
			return parsed, nil
		}
	}

	// Try partial match against managed repos as last resort
	slug, partialErr := discovery.ResolveRepoPartial(input)
	if partialErr == nil {
		parts := strings.SplitN(slug, "/", 2)
		return &ParsedInput{
			Owner: parts[0],
			Repo:  parts[1],
		}, nil
	}

	// Return the partial match error if it's more informative (e.g., "matches multiple repos")
	if strings.Contains(partialErr.Error(), "matches multiple repos") {
		return nil, partialErr
	}

	return nil, fmt.Errorf("invalid input format: %s (expected URL, owner/repo, or partial match)", input)
}

func parsePathLikeInput(input string) (*ParsedInput, bool) {
	p := strings.TrimSpace(input)
	if p == "" {
		return nil, false
	}

	p = filepath.ToSlash(p)
	p = path.Clean(p)

	rawParts := strings.Split(p, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}

	for i, part := range parts {
		itemType, ok := ItemTypeFromDirName(part)
		if !ok {
			continue
		}
		if i < 2 || i+1 >= len(parts) {
			continue
		}

		numberPart := strings.TrimSuffix(parts[i+1], ".md")
		number, err := strconv.Atoi(numberPart)
		if err != nil {
			continue
		}

		return &ParsedInput{
			Owner:    parts[i-2],
			Repo:     parts[i-1],
			Number:   number,
			ItemType: itemType,
		}, true
	}

	return nil, false
}

// Query executes a GraphQL query.
func (c *Client) Query(query string, variables map[string]interface{}, response interface{}) error {
	return c.gql.Do(query, variables, response)
}
