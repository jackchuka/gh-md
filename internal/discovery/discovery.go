package discovery

import (
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
