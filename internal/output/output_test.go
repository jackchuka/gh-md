package output

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestFormatTime(t *testing.T) {
	validTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name   string
		t      *time.Time
		format string
		want   string
	}{
		{
			name:   "nil time",
			t:      nil,
			format: TimestampDisplay,
			want:   "-",
		},
		{
			name:   "zero time",
			t:      &zeroTime,
			format: TimestampDisplay,
			want:   "-",
		},
		{
			name:   "valid time with display format",
			t:      &validTime,
			format: TimestampDisplay,
			want:   validTime.Local().Format(TimestampDisplay),
		},
		{
			name:   "valid time with ISO format",
			t:      &validTime,
			format: TimestampISO,
			want:   validTime.Local().Format(TimestampISO),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.t, tt.format)
			if got != tt.want {
				t.Errorf("FormatTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Format
	}{
		{
			name:  "json",
			input: "json",
			want:  FormatJSON,
		},
		{
			name:  "yaml",
			input: "yaml",
			want:  FormatYAML,
		},
		{
			name:  "text",
			input: "text",
			want:  FormatText,
		},
		{
			name:  "empty string defaults to text",
			input: "",
			want:  FormatText,
		},
		{
			name:  "unknown defaults to text",
			input: "xml",
			want:  FormatText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// mockCommandOutput implements CommandOutput for testing.
type mockCommandOutput struct {
	out io.Writer
	err io.Writer
}

func (m *mockCommandOutput) OutOrStdout() io.Writer { return m.out }
func (m *mockCommandOutput) ErrOrStderr() io.Writer { return m.err }

func TestPrinter_IsStructured(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		want   bool
	}{
		{
			name:   "json format",
			format: FormatJSON,
			want:   true,
		},
		{
			name:   "yaml format",
			format: FormatYAML,
			want:   true,
		},
		{
			name:   "text format",
			format: FormatText,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &mockCommandOutput{out: &buf, err: &buf}
			p := NewPrinter(cmd).WithFormat(tt.format)

			got := p.IsStructured()
			if got != tt.want {
				t.Errorf("IsStructured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		data    any
		wantErr bool
	}{
		{
			name:    "simple struct",
			data:    testStruct{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name:    "slice of structs",
			data:    []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			wantErr: false,
		},
		{
			name:    "nil",
			data:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := JSONOutput(&buf, tt.data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("JSONOutput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify it's valid JSON
				var parsed any
				if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
					t.Errorf("JSONOutput() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestYAMLOutput(t *testing.T) {
	type testStruct struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	tests := []struct {
		name    string
		data    any
		wantErr bool
	}{
		{
			name:    "simple struct",
			data:    testStruct{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name:    "slice of structs",
			data:    []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := YAMLOutput(&buf, tt.data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("YAMLOutput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify it's valid YAML
				var parsed any
				if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
					t.Errorf("YAMLOutput() produced invalid YAML: %v", err)
				}
			}
		})
	}
}

func TestPrinter_Print(t *testing.T) {
	var buf bytes.Buffer
	cmd := &mockCommandOutput{out: &buf, err: &buf}
	p := NewPrinter(cmd)

	p.Print("hello world")

	got := buf.String()
	want := "hello world\n"
	if got != want {
		t.Errorf("Print() wrote %q, want %q", got, want)
	}
}

func TestPrinter_Printf(t *testing.T) {
	var buf bytes.Buffer
	cmd := &mockCommandOutput{out: &buf, err: &buf}
	p := NewPrinter(cmd)

	p.Printf("hello %s %d", "world", 42)

	got := buf.String()
	want := "hello world 42"
	if got != want {
		t.Errorf("Printf() wrote %q, want %q", got, want)
	}
}

func TestPrinter_Errorf(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	cmd := &mockCommandOutput{out: &outBuf, err: &errBuf}
	p := NewPrinter(cmd)

	p.Errorf("error: %s", "something went wrong")

	if outBuf.Len() > 0 {
		t.Error("Errorf() wrote to stdout, should only write to stderr")
	}

	got := errBuf.String()
	want := "error: something went wrong"
	if got != want {
		t.Errorf("Errorf() wrote %q to stderr, want %q", got, want)
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	cmd := &mockCommandOutput{out: &buf, err: &buf}
	p := NewPrinter(cmd)

	table := p.NewTable("NAME", "VALUE")
	table.Row("foo", "1")
	table.Row("bar", "2")
	if err := table.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	got := buf.String()

	// Check headers and rows are present
	if !strings.Contains(got, "NAME") || !strings.Contains(got, "VALUE") {
		t.Errorf("Table output missing headers: %q", got)
	}
	if !strings.Contains(got, "foo") || !strings.Contains(got, "bar") {
		t.Errorf("Table output missing rows: %q", got)
	}
}
