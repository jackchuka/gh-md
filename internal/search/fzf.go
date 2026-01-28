package search

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Action represents an action to perform on a selected item.
type Action string

const (
	ActionOpenEditor  Action = "editor"
	ActionPush        Action = "push"
	ActionViewBrowser Action = "browser"
	ActionCopyPath    Action = "copy"
	ActionPullFresh   Action = "pull"
	ActionCancel      Action = "cancel"
)

// CheckFZFInstalled verifies that fzf is available in PATH.
func CheckFZFInstalled() error {
	_, err := exec.LookPath("fzf")
	if err != nil {
		return fmt.Errorf("fzf not found in PATH. Install it with:\n  brew install fzf (macOS)\n  apt install fzf (Debian/Ubuntu)\n  https://github.com/junegunn/fzf#installation")
	}
	return nil
}

// RunSelector opens fzf with the given items and returns the selected item.
// Returns nil if the user cancels (Esc).
func RunSelector(items []Item, query string, sortBy SortField) (*Item, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to search")
	}

	sortItems(items, sortBy)

	// Build the input for fzf with file paths embedded
	var input strings.Builder
	for i, item := range items {
		// Format: filepath|owner/repo|#number|type|[state]|title
		line := fmt.Sprintf("%s\t%s/%s\t#%d\t%s\t[%s]\t%s",
			item.FilePath,
			item.Owner, item.Repo,
			item.Number,
			item.Type,
			item.State,
			item.Title,
		)
		input.WriteString(line)
		if i < len(items)-1 {
			input.WriteString("\n")
		}
	}

	// Build fzf command with preview using the first field (filepath)
	args := []string{
		"--ansi",
		"--delimiter", "\t",
		"--with-nth", "2..", // Hide the filepath column from display
		"--preview", "cat {1}", // Preview using the first field (filepath)
		"--preview-window", "right:50%:wrap:hidden", // Hidden by default
		"--bind", "ctrl-p:toggle-preview", // Toggle with ctrl-p
		"--header", "Search gh-md files (Enter=select, Ctrl-P=preview, Esc=cancel)",
	}

	if query != "" {
		args = append(args, "--query", query)
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// Exit code 130 = Esc/Ctrl-C, not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return nil, nil
		}
		// Exit code 1 = no match, not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}

	// Parse the selected line
	selected := strings.TrimSpace(stdout.String())
	if selected == "" {
		return nil, nil
	}

	// Extract the filepath from the first column and find the matching item
	parts := strings.SplitN(selected, "\t", 2)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid selection")
	}

	filePath := parts[0]
	for i := range items {
		if items[i].FilePath == filePath {
			return &items[i], nil
		}
	}

	return nil, fmt.Errorf("selected item not found")
}

// RunActionMenu shows a menu of actions for the selected item.
func RunActionMenu(item *Item) (Action, error) {
	actions := []struct {
		action Action
		label  string
	}{
		{ActionOpenEditor, "Open in $EDITOR"},
		{ActionPush, "Push changes to GitHub"},
		{ActionViewBrowser, "View in browser"},
		{ActionCopyPath, "Copy file path"},
		{ActionPullFresh, "Pull fresh from GitHub"},
		{ActionCancel, "Cancel"},
	}

	var input strings.Builder
	for _, a := range actions {
		input.WriteString(a.label)
		input.WriteString("\n")
	}

	header := fmt.Sprintf("Action for %s/%s #%d: %s", item.Owner, item.Repo, item.Number, item.Title)
	if len(header) > 80 {
		header = header[:77] + "..."
	}

	cmd := exec.Command("fzf",
		"--header", header,
		"--no-preview",
		"--height", "~10",
	)
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// Exit code 130 = Esc/Ctrl-C
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return ActionCancel, nil
		}
		// Exit code 1 = no match
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return ActionCancel, nil
		}
		return ActionCancel, err
	}

	selected := strings.TrimSpace(stdout.String())
	for _, a := range actions {
		if a.label == selected {
			return a.action, nil
		}
	}

	return ActionCancel, nil
}
