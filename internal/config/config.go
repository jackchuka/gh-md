package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultRootDir = ".gh-md"
	EnvRootDir     = "GH_MD_ROOT"
)

// GetRootDir returns the root directory for gh-md storage.
// It checks GH_MD_ROOT env var first, then defaults to ~/.gh-md/
func GetRootDir() (string, error) {
	if root := os.Getenv(EnvRootDir); root != "" {
		return root, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, DefaultRootDir), nil
}

// GetRepoDir returns the directory path for a specific repo.
// Format: <root>/<owner>/repos/<repo>/
func GetRepoDir(owner, repo string) (string, error) {
	root, err := GetRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, owner, "repos", repo), nil
}

// GetIssuesDir returns the issues directory for a repo.
func GetIssuesDir(owner, repo string) (string, error) {
	repoDir, err := GetRepoDir(owner, repo)
	if err != nil {
		return "", err
	}

	return filepath.Join(repoDir, "issues"), nil
}

// GetPullsDir returns the pulls directory for a repo.
func GetPullsDir(owner, repo string) (string, error) {
	repoDir, err := GetRepoDir(owner, repo)
	if err != nil {
		return "", err
	}

	return filepath.Join(repoDir, "pulls"), nil
}

// GetDiscussionsDir returns the discussions directory for a repo.
func GetDiscussionsDir(owner, repo string) (string, error) {
	repoDir, err := GetRepoDir(owner, repo)
	if err != nil {
		return "", err
	}

	return filepath.Join(repoDir, "discussions"), nil
}
