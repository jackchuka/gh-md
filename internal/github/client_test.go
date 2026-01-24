package github

import "testing"

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
