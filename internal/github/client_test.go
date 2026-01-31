package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackchuka/gh-md/internal/config"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ParsedInput
		wantErr bool
	}{
		{
			name:  "owner/repo",
			input: "owner/repo",
			want: ParsedInput{
				Owner: "owner",
				Repo:  "repo",
			},
		},
		{
			name:  "owner/repo with dot-slash",
			input: "./owner/repo",
			want: ParsedInput{
				Owner: "owner",
				Repo:  "repo",
			},
		},
		{
			name:  "issue short path",
			input: "owner/repo/issues/123",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				ItemType: ItemTypeIssue,
			},
		},
		{
			name:  "issue short path with .md",
			input: "owner/repo/issues/123.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				ItemType: ItemTypeIssue,
			},
		},
		{
			name:  "issue short path with .md and dot-slash",
			input: "./owner/repo/issues/123.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				ItemType: ItemTypeIssue,
			},
		},
		{
			name:  "issue url with .md",
			input: "https://github.com/owner/repo/issues/123.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   123,
				ItemType: ItemTypeIssue,
			},
		},
		{
			name:  "pr short path (pull)",
			input: "owner/repo/pull/456",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   456,
				ItemType: ItemTypePullRequest,
			},
		},
		{
			name:  "pr short path (pulls) with .md",
			input: "owner/repo/pulls/456.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   456,
				ItemType: ItemTypePullRequest,
			},
		},
		{
			name:  "path-like issue file",
			input: "some/dir/owner/repo/issues/321.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   321,
				ItemType: ItemTypeIssue,
			},
		},
		{
			name:  "discussion short path with .md",
			input: "owner/repo/discussions/789.md",
			want: ParsedInput{
				Owner:    "owner",
				Repo:     "repo",
				Number:   789,
				ItemType: ItemTypeDiscussion,
			},
		},
		{
			name:    "invalid",
			input:   "owner/repo/issues/not-a-number",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseInput error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if *got != tt.want {
				t.Fatalf("ParseInput(%q) = %#v, want %#v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseInput_PartialMatch(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	// Set up a managed repo
	repoDir := filepath.Join(root, "jackchuka", "gh-md")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	metaPath := filepath.Join(repoDir, ".gh-md-meta.yaml")
	if err := os.WriteFile(metaPath, []byte("sync: {}"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    ParsedInput
		wantErr bool
	}{
		{
			name:  "partial match resolves to full repo",
			input: "gh-md",
			want: ParsedInput{
				Owner: "jackchuka",
				Repo:  "gh-md",
			},
		},
		{
			name:  "partial match by owner",
			input: "jackchuka",
			want: ParsedInput{
				Owner: "jackchuka",
				Repo:  "gh-md",
			},
		},
		{
			name:  "exact owner/repo still works",
			input: "jackchuka/gh-md",
			want: ParsedInput{
				Owner: "jackchuka",
				Repo:  "gh-md",
			},
		},
		{
			name:    "no match returns error",
			input:   "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if *got != tt.want {
				t.Fatalf("ParseInput(%q) = %#v, want %#v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseInput_PartialMatchMultiple(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	// Set up repos that will both match "md"
	repos := []struct {
		owner, repo string
	}{
		{"jackchuka", "gh-md"},
		{"acme", "md-tools"},
	}

	for _, r := range repos {
		repoDir := filepath.Join(root, r.owner, r.repo)
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		metaPath := filepath.Join(repoDir, ".gh-md-meta.yaml")
		if err := os.WriteFile(metaPath, []byte("sync: {}"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	_, err := ParseInput("md")
	if err == nil {
		t.Fatal("ParseInput(\"md\") expected error for multiple matches, got nil")
	}

	// Should show the helpful error about multiple matches
	if !strings.Contains(err.Error(), "matches multiple repos") {
		t.Errorf("error = %q, want to contain 'matches multiple repos'", err.Error())
	}
}
