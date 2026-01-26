package github

import "testing"

func TestItemTypeFromDirName(t *testing.T) {
	tests := []struct {
		name   string
		dir    string
		want   ItemType
		wantOK bool
	}{
		{name: "issue singular", dir: "issue", want: ItemTypeIssue, wantOK: true},
		{name: "issues plural", dir: "issues", want: ItemTypeIssue, wantOK: true},
		{name: "pull singular", dir: "pull", want: ItemTypePullRequest, wantOK: true},
		{name: "pulls plural", dir: "pulls", want: ItemTypePullRequest, wantOK: true},
		{name: "discussion singular", dir: "discussion", want: ItemTypeDiscussion, wantOK: true},
		{name: "discussions plural", dir: "discussions", want: ItemTypeDiscussion, wantOK: true},
		{name: "uppercase ISSUES", dir: "ISSUES", want: ItemTypeIssue, wantOK: true},
		{name: "mixed case Pull", dir: "Pull", want: ItemTypePullRequest, wantOK: true},
		{name: "invalid", dir: "commits", want: "", wantOK: false},
		{name: "empty", dir: "", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ItemTypeFromDirName(tt.dir)
			if ok != tt.wantOK {
				t.Errorf("ItemTypeFromDirName(%q) ok = %v, want %v", tt.dir, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("ItemTypeFromDirName(%q) = %q, want %q", tt.dir, got, tt.want)
			}
		})
	}
}

func TestItemType_DirName(t *testing.T) {
	tests := []struct {
		name   string
		t      ItemType
		want   string
		wantOK bool
	}{
		{name: "issue", t: ItemTypeIssue, want: "issues", wantOK: true},
		{name: "pull request", t: ItemTypePullRequest, want: "pulls", wantOK: true},
		{name: "discussion", t: ItemTypeDiscussion, want: "discussions", wantOK: true},
		{name: "invalid", t: ItemType("invalid"), want: "", wantOK: false},
		{name: "empty", t: ItemType(""), want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.t.DirName()
			if ok != tt.wantOK {
				t.Errorf("ItemType(%q).DirName() ok = %v, want %v", tt.t, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("ItemType(%q).DirName() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestItemType_URLSegment(t *testing.T) {
	tests := []struct {
		name   string
		t      ItemType
		want   string
		wantOK bool
	}{
		{name: "issue", t: ItemTypeIssue, want: "issues", wantOK: true},
		{name: "pull request", t: ItemTypePullRequest, want: "pull", wantOK: true},
		{name: "discussion", t: ItemTypeDiscussion, want: "discussions", wantOK: true},
		{name: "invalid", t: ItemType("invalid"), want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.t.URLSegment()
			if ok != tt.wantOK {
				t.Errorf("ItemType(%q).URLSegment() ok = %v, want %v", tt.t, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("ItemType(%q).URLSegment() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestItemType_ListLabel(t *testing.T) {
	tests := []struct {
		name   string
		t      ItemType
		want   string
		wantOK bool
	}{
		{name: "issue", t: ItemTypeIssue, want: "issue", wantOK: true},
		{name: "pull request", t: ItemTypePullRequest, want: "pr", wantOK: true},
		{name: "discussion", t: ItemTypeDiscussion, want: "discussion", wantOK: true},
		{name: "invalid", t: ItemType("invalid"), want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.t.ListLabel()
			if ok != tt.wantOK {
				t.Errorf("ItemType(%q).ListLabel() ok = %v, want %v", tt.t, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("ItemType(%q).ListLabel() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestItemType_Display(t *testing.T) {
	tests := []struct {
		name string
		t    ItemType
		want string
	}{
		{name: "issue", t: ItemTypeIssue, want: "issue"},
		{name: "pull request", t: ItemTypePullRequest, want: "PR"},
		{name: "discussion", t: ItemTypeDiscussion, want: "discussion"},
		{name: "unknown", t: ItemType("unknown"), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.Display()
			if got != tt.want {
				t.Errorf("ItemType(%q).Display() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestItemType_DisplayPlural(t *testing.T) {
	tests := []struct {
		name string
		t    ItemType
		want string
	}{
		{name: "issue", t: ItemTypeIssue, want: "issues"},
		{name: "pull request", t: ItemTypePullRequest, want: "pull requests"},
		{name: "discussion", t: ItemTypeDiscussion, want: "discussions"},
		{name: "unknown", t: ItemType("unknown"), want: "unknowns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.DisplayPlural()
			if got != tt.want {
				t.Errorf("ItemType(%q).DisplayPlural() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}
