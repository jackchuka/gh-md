package github

import "strings"

// ItemTypeFromDirName converts a local directory name or URL path segment into an ItemType.
// Accepted values: "issue", "issues", "pull", "pulls", "discussion", "discussions".
func ItemTypeFromDirName(dir string) (ItemType, bool) {
	switch strings.ToLower(dir) {
	case "issue", "issues":
		return ItemTypeIssue, true
	case "pull", "pulls":
		return ItemTypePullRequest, true
	case "discussion", "discussions":
		return ItemTypeDiscussion, true
	default:
		return "", false
	}
}

// DirName returns the local storage directory name for an item type.
func (t ItemType) DirName() (string, bool) {
	switch t {
	case ItemTypeIssue:
		return "issues", true
	case ItemTypePullRequest:
		return "pulls", true
	case ItemTypeDiscussion:
		return "discussions", true
	default:
		return "", false
	}
}

// URLSegment returns the GitHub web URL path segment for an item type.
func (t ItemType) URLSegment() (string, bool) {
	switch t {
	case ItemTypeIssue:
		return "issues", true
	case ItemTypePullRequest:
		return "pull", true
	case ItemTypeDiscussion:
		return "discussions", true
	default:
		return "", false
	}
}

// ListLabel returns a short label used for displaying an item type in lists.
func (t ItemType) ListLabel() (string, bool) {
	switch t {
	case ItemTypeIssue:
		return "issue", true
	case ItemTypePullRequest:
		return "pr", true
	case ItemTypeDiscussion:
		return "discussion", true
	default:
		return "", false
	}
}
