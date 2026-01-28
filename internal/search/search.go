package search

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cli/go-gh/v2"
	"github.com/google/cel-go/cel"
	"github.com/jackchuka/gh-md/internal/parser"
)

// SortField specifies how to sort items.
type SortField string

const (
	SortUpdated SortField = "updated" // Latest updated first
	SortCreated SortField = "created" // Latest created first
	SortNumber  SortField = "number"  // Highest number first
)

// Item represents a searchable item from local files.
type Item struct {
	FilePath string
	Owner    string
	Repo     string
	Number   int
	Type     string // "issue", "pr", "discussion"
	State    string // "open", "closed", "merged"
	Title    string
	URL      string
	Created  time.Time
	Updated  time.Time
}

// Filters specifies which items to include in search results.
type Filters struct {
	Repo        string // "owner/repo" format, empty = all repos
	Issues      bool
	PRs         bool
	Discussions bool
}

// DiscoverLocalFiles walks the gh-md root directory and returns all matching items.
func DiscoverLocalFiles(filters Filters) ([]Item, error) {
	// If no type filters are set, include all types
	includeAll := !filters.Issues && !filters.PRs && !filters.Discussions

	var items []Item

	err := parser.WalkParsedFiles(parser.WalkFilters{Repo: filters.Repo}, func(parsed *parser.ParsedFile) error {
		itemType := "unknown"
		if label, ok := parsed.ItemType.ListLabel(); ok {
			itemType = label
		}

		// Apply type filters
		if !includeAll {
			switch itemType {
			case "issue":
				if !filters.Issues {
					return nil
				}
			case "pr":
				if !filters.PRs {
					return nil
				}
			case "discussion":
				if !filters.Discussions {
					return nil
				}
			}
		}

		url := ""
		if seg, ok := parsed.ItemType.URLSegment(); ok {
			url = fmt.Sprintf("https://github.com/%s/%s/%s/%d", parsed.Owner, parsed.Repo, seg, parsed.Number)
		}

		item := Item{
			FilePath: parsed.FilePath,
			Owner:    parsed.Owner,
			Repo:     parsed.Repo,
			Number:   parsed.Number,
			Type:     itemType,
			State:    strings.ToLower(parsed.State),
			Title:    parsed.Title,
			URL:      url,
			Created:  parsed.Created,
			Updated:  parsed.Updated,
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetCurrentUser returns the current GitHub username using gh CLI.
func GetCurrentUser() (string, error) {
	stdout, _, err := gh.Exec("api", "user", "--jq", ".login")
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w (is gh CLI authenticated?)", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// DiscoverItems discovers all items matching the CEL filter.
// If repo is empty, searches all repos. If repo is "owner/repo", only searches that repo.
func DiscoverItems(prg cel.Program, username string, repo string) ([]Item, error) {
	var items []Item

	filters := parser.WalkFilters{Repo: repo}

	err := parser.WalkParsedFiles(filters, func(parsed *parser.ParsedFile) error {
		// Determine item type label
		itemType := "unknown"
		if label, ok := parsed.ItemType.ListLabel(); ok {
			itemType = label
		}

		// Build URL
		url := ""
		if seg, ok := parsed.ItemType.URLSegment(); ok {
			url = fmt.Sprintf("https://github.com/%s/%s/%s/%d", parsed.Owner, parsed.Repo, seg, parsed.Number)
		}

		// Ensure slices are not nil (CEL requires non-nil lists)
		assigned := parsed.Assignees
		if assigned == nil {
			assigned = []string{}
		}
		reviewers := parsed.Reviewers
		if reviewers == nil {
			reviewers = []string{}
		}
		labels := parsed.Labels
		if labels == nil {
			labels = []string{}
		}

		// Build CEL variables map
		vars := map[string]any{
			"user":      username,
			"item_type": itemType,
			"state":     strings.ToLower(parsed.State),
			"title":     parsed.Title,
			"body":      parsed.Body,
			"author":    parsed.Author,
			"assigned":  assigned,
			"reviewers": reviewers,
			"labels":    labels,
			"created":   parsed.Created,
			"updated":   parsed.Updated,
			"owner":     parsed.Owner,
			"repo":      parsed.Repo,
			"number":    parsed.Number,
		}

		// Evaluate the CEL filter
		match, err := EvaluateFilter(prg, vars)
		if err != nil {
			// Skip items that fail evaluation (e.g., missing fields)
			return nil
		}

		if match {
			items = append(items, Item{
				FilePath: parsed.FilePath,
				Owner:    parsed.Owner,
				Repo:     parsed.Repo,
				Number:   parsed.Number,
				Type:     itemType,
				State:    strings.ToLower(parsed.State),
				Title:    parsed.Title,
				URL:      url,
				Created:  parsed.Created,
				Updated:  parsed.Updated,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

func sortItems(items []Item, field SortField) {
	sort.Slice(items, func(i, j int) bool {
		switch field {
		case SortCreated:
			return items[i].Created.After(items[j].Created)
		case SortNumber:
			return items[i].Number > items[j].Number
		default: // SortUpdated
			return items[i].Updated.After(items[j].Updated)
		}
	})
}
