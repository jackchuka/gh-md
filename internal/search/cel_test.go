package search

import (
	"testing"
	"time"
)

func TestNewCELEnv(t *testing.T) {
	env, err := NewCELEnv()
	if err != nil {
		t.Fatalf("NewCELEnv() error = %v", err)
	}
	if env == nil {
		t.Fatal("NewCELEnv() returned nil")
	}
}

func TestCompileCELFilter(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "valid simple expression",
			expr:    `state == "open"`,
			wantErr: false,
		},
		{
			name:    "valid list contains",
			expr:    `"bug" in labels`,
			wantErr: false,
		},
		{
			name:    "valid compound expression",
			expr:    `state == "open" && author == user`,
			wantErr: false,
		},
		{
			name:    "valid time comparison",
			expr:    `updated > now - duration("24h")`,
			wantErr: false,
		},
		{
			name:    "valid string contains",
			expr:    `title.contains("fix")`,
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			expr:    `state ==`,
			wantErr: true,
		},
		{
			name:    "invalid variable",
			expr:    `undefined_var == "test"`,
			wantErr: true,
		},
		{
			name:    "invalid operator",
			expr:    `state === "open"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prg, err := CompileCELFilter(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompileCELFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && prg == nil {
				t.Error("CompileCELFilter() returned nil program")
			}
		})
	}
}

func TestEvaluateFilter(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    bool
		wantErr bool
	}{
		{
			name: "state equals open",
			expr: `state == "open"`,
			vars: map[string]any{
				"state": "open",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "state equals open - false",
			expr: `state == "open"`,
			vars: map[string]any{
				"state": "closed",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "author equals user",
			expr: `author == user`,
			vars: map[string]any{
				"author": "octocat",
				"user":   "octocat",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "label in labels",
			expr: `"bug" in labels`,
			vars: map[string]any{
				"labels": []string{"bug", "critical"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "label not in labels",
			expr: `"feature" in labels`,
			vars: map[string]any{
				"labels": []string{"bug", "critical"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "compound and expression",
			expr: `state == "open" && "bug" in labels`,
			vars: map[string]any{
				"state":  "open",
				"labels": []string{"bug"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "compound or expression",
			expr: `state == "closed" || state == "merged"`,
			vars: map[string]any{
				"state": "merged",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "title contains",
			expr: `title.contains("fix")`,
			vars: map[string]any{
				"title": "fix: resolve bug",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "number comparison",
			expr: `number > 100`,
			vars: map[string]any{
				"number": int64(150),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "assignee in list",
			expr: `"dev1" in assigned`,
			vars: map[string]any{
				"assigned": []string{"dev1", "dev2"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "reviewer in list",
			expr: `"reviewer1" in reviewers`,
			vars: map[string]any{
				"reviewers": []string{"reviewer1"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "owner and repo",
			expr: `owner == "github" && repo == "docs"`,
			vars: map[string]any{
				"owner": "github",
				"repo":  "docs",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "item type check",
			expr: `item_type == "issue"`,
			vars: map[string]any{
				"item_type": "issue",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "time comparison - recent",
			expr: `created > now - duration("720h")`, // 30 days
			vars: map[string]any{
				"created": time.Now().Add(-24 * time.Hour), // 1 day ago
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "time comparison - old",
			expr: `updated > now - duration("24h")`,
			vars: map[string]any{
				"updated": time.Now().Add(-48 * time.Hour), // 2 days ago
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "empty labels list",
			expr: `size(labels) == 0`,
			vars: map[string]any{
				"labels": []string{},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prg, err := CompileCELFilter(tt.expr)
			if err != nil {
				t.Fatalf("CompileCELFilter() error = %v", err)
			}

			got, err := EvaluateFilter(prg, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("EvaluateFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateFilter_NonBooleanResult(t *testing.T) {
	// Expression that returns a string instead of boolean
	expr := `state`
	prg, err := CompileCELFilter(expr)
	if err != nil {
		t.Fatalf("CompileCELFilter() error = %v", err)
	}

	vars := map[string]any{
		"state": "open",
	}

	_, err = EvaluateFilter(prg, vars)
	if err == nil {
		t.Error("EvaluateFilter() error = nil, want error for non-boolean result")
	}
}
