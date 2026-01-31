package discovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/meta"
)

func TestManagedRepo_Slug(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		repo  string
		want  string
	}{
		{
			name:  "simple",
			owner: "octocat",
			repo:  "hello-world",
			want:  "octocat/hello-world",
		},
		{
			name:  "with org",
			owner: "github",
			repo:  "docs",
			want:  "github/docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ManagedRepo{Owner: tt.owner, Repo: tt.repo}
			got := r.Slug()
			if got != tt.want {
				t.Errorf("Slug() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManagedRepo_LastSyncTime(t *testing.T) {
	ts1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	ts3 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		meta *meta.Meta
		want *time.Time
	}{
		{
			name: "nil meta",
			meta: nil,
			want: nil,
		},
		{
			name: "nil sync",
			meta: &meta.Meta{Sync: nil},
			want: nil,
		},
		{
			name: "only issues timestamp",
			meta: &meta.Meta{
				Sync: &meta.SyncTimestamps{Issues: &ts1},
			},
			want: &ts1,
		},
		{
			name: "multiple timestamps returns latest",
			meta: &meta.Meta{
				Sync: &meta.SyncTimestamps{
					Issues:      &ts1,
					Pulls:       &ts3,
					Discussions: &ts2,
				},
			},
			want: &ts3,
		},
		{
			name: "all nil timestamps",
			meta: &meta.Meta{
				Sync: &meta.SyncTimestamps{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ManagedRepo{
				Owner: "test",
				Repo:  "repo",
				Meta:  tt.meta,
			}
			got := r.LastSyncTime()

			if tt.want == nil {
				if got != nil {
					t.Errorf("LastSyncTime() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("LastSyncTime() = nil, want %v", tt.want)
				return
			}

			if !got.Equal(*tt.want) {
				t.Errorf("LastSyncTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverManagedRepos(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	t.Run("empty root", func(t *testing.T) {
		repos, err := DiscoverManagedRepos()
		if err != nil {
			t.Fatalf("DiscoverManagedRepos() error = %v", err)
		}
		if len(repos) != 0 {
			t.Errorf("DiscoverManagedRepos() = %d repos, want 0", len(repos))
		}
	})

	t.Run("single repo with meta", func(t *testing.T) {
		// Create repo with meta file
		repoDir := filepath.Join(root, "owner1", "repo1")
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		metaContent := `sync:
  issues: 2026-01-15T10:00:00Z
`
		metaPath := filepath.Join(repoDir, metaFile)
		if err := os.WriteFile(metaPath, []byte(metaContent), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		repos, err := DiscoverManagedRepos()
		if err != nil {
			t.Fatalf("DiscoverManagedRepos() error = %v", err)
		}

		if len(repos) != 1 {
			t.Fatalf("DiscoverManagedRepos() = %d repos, want 1", len(repos))
		}

		if repos[0].Owner != "owner1" || repos[0].Repo != "repo1" {
			t.Errorf("repo = %s/%s, want owner1/repo1", repos[0].Owner, repos[0].Repo)
		}
	})

	t.Run("multiple repos sorted by slug", func(t *testing.T) {
		// Create additional repos
		repos := []struct {
			owner, repo string
		}{
			{"zebra", "last"},
			{"alpha", "first"},
		}

		for _, r := range repos {
			repoDir := filepath.Join(root, r.owner, r.repo)
			if err := os.MkdirAll(repoDir, 0755); err != nil {
				t.Fatalf("MkdirAll failed: %v", err)
			}
			metaPath := filepath.Join(repoDir, metaFile)
			if err := os.WriteFile(metaPath, []byte("sync: {}"), 0644); err != nil {
				t.Fatalf("WriteFile failed: %v", err)
			}
		}

		discovered, err := DiscoverManagedRepos()
		if err != nil {
			t.Fatalf("DiscoverManagedRepos() error = %v", err)
		}

		// Should be sorted: alpha/first, owner1/repo1, zebra/last
		if len(discovered) < 3 {
			t.Fatalf("DiscoverManagedRepos() = %d repos, want at least 3", len(discovered))
		}

		if discovered[0].Slug() != "alpha/first" {
			t.Errorf("first repo = %s, want alpha/first", discovered[0].Slug())
		}
	})

	t.Run("repo without meta file skipped", func(t *testing.T) {
		// Create repo directory without meta file
		repoDir := filepath.Join(root, "no-meta", "repo")
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		discovered, err := DiscoverManagedRepos()
		if err != nil {
			t.Fatalf("DiscoverManagedRepos() error = %v", err)
		}

		// Should not include no-meta/repo
		for _, r := range discovered {
			if r.Owner == "no-meta" {
				t.Errorf("discovered repo without meta file: %s", r.Slug())
			}
		}
	})
}

func TestResolveRepoPartial(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	// Set up test repos
	repos := []struct {
		owner, repo string
	}{
		{"jackchuka", "gh-md"},
		{"acme", "gh-md-tools"},
		{"org", "awesome"},
	}

	for _, r := range repos {
		repoDir := filepath.Join(root, r.owner, r.repo)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		metaPath := filepath.Join(repoDir, metaFile)
		if err := os.WriteFile(metaPath, []byte("sync: {}"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	tests := []struct {
		name    string
		partial string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "single match by repo name",
			partial: "awesome",
			want:    "org/awesome",
		},
		{
			name:    "single match by owner name",
			partial: "jackchuka",
			want:    "jackchuka/gh-md",
		},
		{
			name:    "case insensitive match",
			partial: "AWESOME",
			want:    "org/awesome",
		},
		{
			name:    "multiple matches returns error",
			partial: "gh-md",
			wantErr: true,
			errMsg:  "matches multiple repos",
		},
		{
			name:    "no matches returns error",
			partial: "nonexistent",
			wantErr: true,
			errMsg:  "no repos match",
		},
		{
			name:    "empty input returns error",
			partial: "",
			wantErr: true,
		},
		{
			name:    "whitespace trimmed",
			partial: "  awesome  ",
			want:    "org/awesome",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveRepoPartial(tt.partial)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveRepoPartial(%q) error = %v, wantErr %v", tt.partial, err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errMsg != "" && err != nil {
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
					}
				}
				return
			}
			if got != tt.want {
				t.Errorf("ResolveRepoPartial(%q) = %q, want %q", tt.partial, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
