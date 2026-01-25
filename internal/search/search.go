package search

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackchuka/gh-md/internal/config"
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
	root, err := config.GetRootDir()
	if err != nil {
		return nil, err
	}

	// If no type filters are set, include all types
	includeAll := !filters.Issues && !filters.PRs && !filters.Discussions

	var items []Item

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Only process .md files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Parse the file to extract metadata
		parsed, err := parser.ParseFile(path)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Apply repo filter
		if filters.Repo != "" {
			repoPath := parsed.Owner + "/" + parsed.Repo
			if repoPath != filters.Repo {
				return nil
			}
		}

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
			FilePath: path,
			Owner:    parsed.Owner,
			Repo:     parsed.Repo,
			Number:   parsed.Number,
			Type:     itemType,
			State:    normalizeState(parsed.State),
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

// normalizeState converts state to a consistent lowercase format.
func normalizeState(state string) string {
	return strings.ToLower(state)
}
