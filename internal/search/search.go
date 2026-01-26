package search

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-md/internal/parser"
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
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}
