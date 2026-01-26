package prune

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackchuka/gh-md/internal/github"
)

func TestPruneResult_RelativePath(t *testing.T) {
	tests := []struct {
		name     string
		result   PruneResult
		wantPath string
	}{
		{
			name: "issue",
			result: PruneResult{
				Path:     "/some/root/owner/repo/issues/123.md",
				ItemType: github.ItemTypeIssue,
				Number:   123,
				Owner:    "owner",
				Repo:     "repo",
			},
			wantPath: filepath.Join("owner", "repo", "issues", "123.md"),
		},
		{
			name: "pull request",
			result: PruneResult{
				Path:     "/some/root/owner/repo/pulls/456.md",
				ItemType: github.ItemTypePullRequest,
				Number:   456,
				Owner:    "owner",
				Repo:     "repo",
			},
			wantPath: filepath.Join("owner", "repo", "pulls", "456.md"),
		},
		{
			name: "discussion",
			result: PruneResult{
				Path:     "/some/root/owner/repo/discussions/789.md",
				ItemType: github.ItemTypeDiscussion,
				Number:   789,
				Owner:    "owner",
				Repo:     "repo",
			},
			wantPath: filepath.Join("owner", "repo", "discussions", "789.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.RelativePath()
			if got != tt.wantPath {
				t.Errorf("RelativePath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestDeleteFiles(t *testing.T) {
	t.Run("delete existing files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		files := []string{
			filepath.Join(tmpDir, "file1.md"),
			filepath.Join(tmpDir, "file2.md"),
			filepath.Join(tmpDir, "file3.md"),
		}

		for _, f := range files {
			if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
				t.Fatalf("WriteFile failed: %v", err)
			}
		}

		// Create PruneResults
		results := make([]PruneResult, len(files))
		for i, f := range files {
			results[i] = PruneResult{
				Path:     f,
				ItemType: github.ItemTypeIssue,
				Number:   i + 1,
				Owner:    "owner",
				Repo:     "repo",
			}
		}

		deleted, err := DeleteFiles(results)
		if err != nil {
			t.Fatalf("DeleteFiles() error = %v", err)
		}

		if deleted != len(files) {
			t.Errorf("DeleteFiles() deleted = %d, want %d", deleted, len(files))
		}

		// Verify files are gone
		for _, f := range files {
			if _, err := os.Stat(f); !os.IsNotExist(err) {
				t.Errorf("file %s still exists after delete", f)
			}
		}
	})

	t.Run("file not exists returns error", func(t *testing.T) {
		results := []PruneResult{
			{
				Path:     "/nonexistent/path/file.md",
				ItemType: github.ItemTypeIssue,
				Number:   1,
				Owner:    "owner",
				Repo:     "repo",
			},
		}

		_, err := DeleteFiles(results)
		if err == nil {
			t.Error("DeleteFiles() error = nil, want error for nonexistent file")
		}
	})

	t.Run("partial delete returns count", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create first file
		existingFile := filepath.Join(tmpDir, "existing.md")
		if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		results := []PruneResult{
			{Path: existingFile, ItemType: github.ItemTypeIssue, Number: 1, Owner: "o", Repo: "r"},
			{Path: "/nonexistent/file.md", ItemType: github.ItemTypeIssue, Number: 2, Owner: "o", Repo: "r"},
		}

		deleted, err := DeleteFiles(results)
		if err == nil {
			t.Error("DeleteFiles() error = nil, want error")
		}
		if deleted != 1 {
			t.Errorf("DeleteFiles() deleted = %d, want 1", deleted)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		deleted, err := DeleteFiles([]PruneResult{})
		if err != nil {
			t.Fatalf("DeleteFiles() error = %v", err)
		}
		if deleted != 0 {
			t.Errorf("DeleteFiles() deleted = %d, want 0", deleted)
		}
	})
}
