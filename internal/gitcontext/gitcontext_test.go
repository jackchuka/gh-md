package gitcontext

import "testing"

func TestIsDefaultBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"master", true},
		{"develop", false},
		{"feature/foo", false},
		{"Main", false},   // case sensitive
		{"MASTER", false}, // case sensitive
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			if got := isDefaultBranch(tt.branch); got != tt.want {
				t.Errorf("isDefaultBranch(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}
