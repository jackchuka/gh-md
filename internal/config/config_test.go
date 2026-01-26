package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRootDir(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		wantEnv bool
	}{
		{
			name:    "env var set",
			envVar:  "/custom/path",
			wantEnv: true,
		},
		{
			name:    "env var empty",
			envVar:  "",
			wantEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(EnvRootDir, tt.envVar)
			} else {
				t.Setenv(EnvRootDir, "")
			}

			got, err := GetRootDir()
			if err != nil {
				t.Fatalf("GetRootDir() error = %v", err)
			}

			if tt.wantEnv {
				if got != tt.envVar {
					t.Errorf("GetRootDir() = %q, want %q", got, tt.envVar)
				}
			} else {
				home, _ := os.UserHomeDir()
				want := filepath.Join(home, DefaultRootDir)
				if got != want {
					t.Errorf("GetRootDir() = %q, want %q", got, want)
				}
			}
		})
	}
}

func TestGetRepoDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRootDir, root)

	tests := []struct {
		name  string
		owner string
		repo  string
		want  string
	}{
		{
			name:  "valid owner/repo",
			owner: "octocat",
			repo:  "hello-world",
			want:  filepath.Join(root, "octocat", "hello-world"),
		},
		{
			name:  "different owner/repo",
			owner: "github",
			repo:  "docs",
			want:  filepath.Join(root, "github", "docs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRepoDir(tt.owner, tt.repo)
			if err != nil {
				t.Fatalf("GetRepoDir() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GetRepoDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetIssuesDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRootDir, root)

	got, err := GetIssuesDir("owner", "repo")
	if err != nil {
		t.Fatalf("GetIssuesDir() error = %v", err)
	}

	want := filepath.Join(root, "owner", "repo", "issues")
	if got != want {
		t.Errorf("GetIssuesDir() = %q, want %q", got, want)
	}
}

func TestGetPullsDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRootDir, root)

	got, err := GetPullsDir("owner", "repo")
	if err != nil {
		t.Fatalf("GetPullsDir() error = %v", err)
	}

	want := filepath.Join(root, "owner", "repo", "pulls")
	if got != want {
		t.Errorf("GetPullsDir() = %q, want %q", got, want)
	}
}

func TestGetDiscussionsDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRootDir, root)

	got, err := GetDiscussionsDir("owner", "repo")
	if err != nil {
		t.Fatalf("GetDiscussionsDir() error = %v", err)
	}

	want := filepath.Join(root, "owner", "repo", "discussions")
	if got != want {
		t.Errorf("GetDiscussionsDir() = %q, want %q", got, want)
	}
}
