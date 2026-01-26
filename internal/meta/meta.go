package meta

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jackchuka/gh-md/internal/config"
	"gopkg.in/yaml.v3"
)

const metaFile = ".gh-md-meta.yaml"

// Meta is the root structure for the metadata file.
// It allows for future extensibility beyond sync timestamps.
type Meta struct {
	Sync *SyncTimestamps `yaml:"sync,omitempty"`
}

// SyncTimestamps stores the last sync timestamps for each item type.
type SyncTimestamps struct {
	Issues      *time.Time `yaml:"issues,omitempty"`
	Pulls       *time.Time `yaml:"pulls,omitempty"`
	Discussions *time.Time `yaml:"discussions,omitempty"`
	// Previous timestamps for --new flag (items updated since previous sync)
	PrevIssues      *time.Time `yaml:"prev_issues,omitempty"`
	PrevPulls       *time.Time `yaml:"prev_pulls,omitempty"`
	PrevDiscussions *time.Time `yaml:"prev_discussions,omitempty"`
}

// Load loads metadata for a repository.
// Returns an empty Meta if the file doesn't exist.
func Load(owner, repo string) (*Meta, error) {
	path, err := metaPath(owner, repo)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Meta{}, nil
		}
		return nil, err
	}

	var meta Meta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// Save saves metadata for a repository with atomic write.
func Save(owner, repo string, meta *Meta) error {
	path, err := metaPath(owner, repo)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func metaPath(owner, repo string) (string, error) {
	repoDir, err := config.GetRepoDir(owner, repo)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoDir, metaFile), nil
}
