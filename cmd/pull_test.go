package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/github"
)

func TestResolvePullInput(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvRootDir, root)

	repoDir := filepath.Join(root, "acme", "widgets")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".gh-md-meta.yaml"), []byte("sync: {}"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	currentUser := func() (string, error) { return "octocat", nil }

	tests := []struct {
		name    string
		input   string
		user    func() (string, error)
		want    github.ParsedInput
		wantErr bool
	}{
		{
			name:  "partial name matching a managed repo",
			input: "widg",
			user:  currentUser,
			want:  github.ParsedInput{Owner: "acme", Repo: "widgets"},
		},
		{
			name:  "bare name with no managed match defaults to current user",
			input: "mdschema",
			user:  currentUser,
			want:  github.ParsedInput{Owner: "octocat", Repo: "mdschema"},
		},
		{
			name:  "owner/repo is used as-is",
			input: "other/mdschema",
			user:  currentUser,
			want:  github.ParsedInput{Owner: "other", Repo: "mdschema"},
		},
		{
			name:    "current user lookup failure is reported",
			input:   "mdschema",
			user:    func() (string, error) { return "", errors.New("not authenticated") },
			wantErr: true,
		},
		{
			name:    "invalid repo name is not guessed",
			input:   "not a repo",
			user:    currentUser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePullInput(tt.input, tt.user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolvePullInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if *got != tt.want {
				t.Fatalf("resolvePullInput(%q) = %#v, want %#v", tt.input, *got, tt.want)
			}
		})
	}
}
