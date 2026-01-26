package prune

import (
	"os"
	"path/filepath"

	"github.com/jackchuka/gh-md/internal/github"
	"github.com/jackchuka/gh-md/internal/parser"
)

// PruneResult represents a file that can be pruned.
type PruneResult struct {
	Path     string
	ItemType github.ItemType
	Number   int
	State    string
	Owner    string
	Repo     string
}

// RelativePath returns the path relative to the gh-md root for display.
func (p *PruneResult) RelativePath() string {
	dirName, _ := p.ItemType.DirName()
	return filepath.Join(p.Owner, p.Repo, dirName, filepath.Base(p.Path))
}

// FindPrunableFiles walks the gh-md root and returns files that should be pruned.
// If repoFilter is non-empty (format: "owner/repo"), only files from that repo are included.
// Prunable files are:
// - Issues with state == "closed"
// - Pull requests with state == "merged" or state == "closed"
// Discussions are never pruned.
func FindPrunableFiles(repoFilter string) ([]PruneResult, error) {
	var results []PruneResult

	err := parser.WalkParsedFiles(parser.WalkFilters{Repo: repoFilter}, func(parsed *parser.ParsedFile) error {
		// Skip if no state information
		if parsed.State == "" {
			return nil
		}

		// Determine if this file should be pruned based on item type and state
		shouldPrune := false
		switch parsed.ItemType {
		case github.ItemTypeIssue, github.ItemTypeDiscussion:
			// Prune closed issues and discussions
			shouldPrune = parsed.State == "closed"
		case github.ItemTypePullRequest:
			// Prune merged or closed PRs
			shouldPrune = parsed.State == "merged" || parsed.State == "closed"
		}

		if shouldPrune {
			results = append(results, PruneResult{
				Path:     parsed.FilePath,
				ItemType: parsed.ItemType,
				Number:   parsed.Number,
				State:    parsed.State,
				Owner:    parsed.Owner,
				Repo:     parsed.Repo,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteFiles deletes the specified files and returns the number of files deleted.
func DeleteFiles(files []PruneResult) (int, error) {
	deleted := 0
	for _, f := range files {
		if err := os.Remove(f.Path); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}
