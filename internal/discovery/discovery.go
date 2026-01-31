package discovery

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/meta"
)

const metaFile = ".gh-md-meta.yaml"

// ManagedRepo represents a discovered managed repository.
type ManagedRepo struct {
	Owner string
	Repo  string
	Path  string
	Meta  *meta.Meta
}

// Slug returns "owner/repo" format.
func (r *ManagedRepo) Slug() string {
	return r.Owner + "/" + r.Repo
}

// LastSyncTime returns the most recent sync timestamp across all item types.
func (r *ManagedRepo) LastSyncTime() *time.Time {
	if r.Meta == nil || r.Meta.Sync == nil {
		return nil
	}

	var latest *time.Time
	times := []*time.Time{
		r.Meta.Sync.Issues,
		r.Meta.Sync.Pulls,
		r.Meta.Sync.Discussions,
	}

	for _, t := range times {
		if t != nil && (latest == nil || t.After(*latest)) {
			latest = t
		}
	}

	return latest
}

// DiscoverManagedRepos scans the gh-md root for all managed repositories.
// A repository is considered "managed" if it has a .gh-md-meta.yaml file.
func DiscoverManagedRepos() ([]ManagedRepo, error) {
	root, err := config.GetRootDir()
	if err != nil {
		return nil, err
	}

	var repos []ManagedRepo

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Only interested in .gh-md-meta.yaml files
		if d.IsDir() || d.Name() != metaFile {
			return nil
		}

		// Extract owner/repo from path: root/owner/repo/.gh-md-meta.yaml
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		parts := strings.Split(filepath.Dir(relPath), string(filepath.Separator))
		if len(parts) != 2 {
			return nil // Not in expected owner/repo structure
		}

		owner, repo := parts[0], parts[1]

		// Load metadata
		md, err := meta.Load(owner, repo)
		if err != nil {
			return nil // Skip repos with corrupted metadata
		}

		repos = append(repos, ManagedRepo{
			Owner: owner,
			Repo:  repo,
			Path:  filepath.Dir(path),
			Meta:  md,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by owner/repo
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Slug() < repos[j].Slug()
	})

	return repos, nil
}

// ResolveRepoPartial attempts to match a partial input against managed repos.
// Returns the matching repo slug ("owner/repo") if exactly one match found.
// Returns an error with suggestions if zero or multiple matches.
func ResolveRepoPartial(partial string) (string, error) {
	partial = strings.TrimSpace(partial)
	if partial == "" {
		return "", fmt.Errorf("empty input")
	}

	repos, err := DiscoverManagedRepos()
	if err != nil {
		return "", err
	}

	partialLower := strings.ToLower(partial)
	var matches []string

	for _, r := range repos {
		slug := r.Slug()
		if strings.Contains(strings.ToLower(slug), partialLower) {
			matches = append(matches, slug)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no repos match %q. Run \"gh md repos\" to see managed repos", partial)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%q matches multiple repos:\n  - %s\nUse full owner/repo format to be specific",
			partial, strings.Join(matches, "\n  - "))
	}
}
