package meta

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackchuka/gh-md/internal/config"
)

func TestLoad(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	t.Run("file not exists", func(t *testing.T) {
		meta, err := Load("nonexistent", "repo")
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}
		if meta == nil {
			t.Fatal("Load() returned nil, want empty Meta")
		}
		if meta.Sync != nil {
			t.Errorf("Load() returned Meta with non-nil Sync, want nil")
		}
	})

	t.Run("valid yaml file", func(t *testing.T) {
		owner, repo := "test", "valid"
		repoDir := filepath.Join(root, owner, repo)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		content := `sync:
  issues: 2026-01-15T10:00:00Z
  pulls: 2026-01-15T10:00:00Z
`
		metaPath := filepath.Join(repoDir, ".gh-md-meta.yaml")
		if err := os.WriteFile(metaPath, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		meta, err := Load(owner, repo)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if meta.Sync == nil {
			t.Fatal("Load() returned Meta with nil Sync")
		}
		if meta.Sync.Issues == nil || !meta.Sync.Issues.Equal(ts) {
			t.Errorf("Load() Issues = %v, want %v", meta.Sync.Issues, ts)
		}
		if meta.Sync.Pulls == nil || !meta.Sync.Pulls.Equal(ts) {
			t.Errorf("Load() Pulls = %v, want %v", meta.Sync.Pulls, ts)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		owner, repo := "test", "invalid"
		repoDir := filepath.Join(root, owner, repo)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		content := `not: valid: yaml: content`
		metaPath := filepath.Join(repoDir, ".gh-md-meta.yaml")
		if err := os.WriteFile(metaPath, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		_, err := Load(owner, repo)
		if err == nil {
			t.Error("Load() error = nil, want error for invalid yaml")
		}
	})
}

func TestSave(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	t.Run("new file", func(t *testing.T) {
		owner, repo := "save", "newfile"
		ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

		meta := &Meta{
			Sync: &SyncTimestamps{
				Issues: &ts,
			},
		}

		if err := Save(owner, repo, meta); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file exists and can be loaded
		loaded, err := Load(owner, repo)
		if err != nil {
			t.Fatalf("Load() after Save() error = %v", err)
		}

		if loaded.Sync == nil || loaded.Sync.Issues == nil {
			t.Fatal("Loaded Meta missing Sync.Issues")
		}
		if !loaded.Sync.Issues.Equal(ts) {
			t.Errorf("Loaded Issues = %v, want %v", loaded.Sync.Issues, ts)
		}
	})

	t.Run("overwrite existing", func(t *testing.T) {
		owner, repo := "save", "overwrite"
		ts1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		ts2 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

		// Save initial
		meta1 := &Meta{Sync: &SyncTimestamps{Issues: &ts1}}
		if err := Save(owner, repo, meta1); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Overwrite
		meta2 := &Meta{Sync: &SyncTimestamps{Issues: &ts2, Pulls: &ts2}}
		if err := Save(owner, repo, meta2); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify
		loaded, err := Load(owner, repo)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if !loaded.Sync.Issues.Equal(ts2) {
			t.Errorf("Issues = %v, want %v", loaded.Sync.Issues, ts2)
		}
		if loaded.Sync.Pulls == nil || !loaded.Sync.Pulls.Equal(ts2) {
			t.Errorf("Pulls = %v, want %v", loaded.Sync.Pulls, ts2)
		}
	})

	t.Run("directory not exists", func(t *testing.T) {
		owner, repo := "new", "deep/nested/repo"
		ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

		meta := &Meta{Sync: &SyncTimestamps{Discussions: &ts}}
		if err := Save(owner, repo, meta); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify directory was created
		repoDir := filepath.Join(root, owner, repo)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			t.Error("Save() did not create directory")
		}
	})
}
