package gitcontext

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// Context holds detected git repository information.
type Context struct {
	Owner  string
	Repo   string
	Branch string
}

// Result holds the resolved action to take based on git context.
type Result struct {
	Owner     string
	Repo      string
	PRNumber  int  // >0 if current branch has an open PR
	IsDefault bool // true if on main/master branch
}

// Detect attempts to detect git context from the current directory.
// Returns an error if not in a git repository or detection fails.
func Detect() (*Context, error) {
	// Use go-gh's repository.Current() which handles:
	// - git remote detection
	// - gh CLI authentication context
	// - Multiple remotes (uses origin or gh-default)
	repo, err := repository.Current()
	if err != nil {
		return nil, fmt.Errorf("not in a GitHub repository: %w", err)
	}

	// Get current branch using git
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(stdout.String())
	if branch == "HEAD" {
		return nil, fmt.Errorf("detached HEAD state - cannot determine branch")
	}

	return &Context{
		Owner:  repo.Owner,
		Repo:   repo.Name,
		Branch: branch,
	}, nil
}

// isDefaultBranch checks if a branch is a default branch (main or master).
func isDefaultBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

// FindPRForBranch finds an open PR for the given branch in the repository.
// Returns 0 if no open PR exists for the branch.
func FindPRForBranch(owner, repo, branch string) (int, error) {
	// Use gh CLI to search for PRs with this head branch
	stdout, _, err := gh.Exec(
		"pr", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--head", branch,
		"--state", "open",
		"--json", "number",
		"--jq", ".[0].number // 0",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to find PR for branch: %w", err)
	}

	numStr := strings.TrimSpace(stdout.String())
	if numStr == "" || numStr == "0" || numStr == "null" {
		return 0, nil
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, nil
	}

	return num, nil
}

// Resolve determines what pull action to take based on the git context.
func (c *Context) Resolve() (*Result, error) {
	result := &Result{
		Owner: c.Owner,
		Repo:  c.Repo,
	}

	// On default branch -> pull all items for this repo
	if isDefaultBranch(c.Branch) {
		result.IsDefault = true
		return result, nil
	}

	// On feature branch -> check for open PR
	prNum, err := FindPRForBranch(c.Owner, c.Repo, c.Branch)
	if err != nil {
		return nil, err
	}

	if prNum == 0 {
		return nil, fmt.Errorf(
			"branch '%s' has no open PR\n\n"+
				"To pull all items for this repo: gh md pull %s/%s\n"+
				"To create a PR first: gh pr create",
			c.Branch, c.Owner, c.Repo,
		)
	}

	result.PRNumber = prNum
	return result, nil
}
