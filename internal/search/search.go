package search

import (
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

		// Determine item type from path
		itemType := detectType(path)

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

		item := Item{
			FilePath: path,
			Owner:    parsed.Owner,
			Repo:     parsed.Repo,
			Number:   parsed.Number,
			Type:     itemType,
			State:    normalizeState(parsed.State),
			Title:    parsed.Title,
			URL:      buildURL(parsed.Owner, parsed.Repo, itemType, parsed.Number),
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

// detectType determines the item type from the file path.
func detectType(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)

	switch base {
	case "issues":
		return "issue"
	case "pulls":
		return "pr"
	case "discussions":
		return "discussion"
	default:
		return "unknown"
	}
}

// normalizeState converts state to a consistent lowercase format.
func normalizeState(state string) string {
	return strings.ToLower(state)
}

// buildURL constructs a GitHub URL for the item.
func buildURL(owner, repo, itemType string, number int) string {
	var typeSegment string
	switch itemType {
	case "issue":
		typeSegment = "issues"
	case "pr":
		typeSegment = "pull"
	case "discussion":
		typeSegment = "discussions"
	default:
		return ""
	}

	return "https://github.com/" + owner + "/" + repo + "/" + typeSegment + "/" + itoa(number)
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
