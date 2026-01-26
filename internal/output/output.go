package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

// Timestamp format constants for consistent display across commands.
const (
	// TimestampDisplay is for human-readable output in terminal.
	TimestampDisplay = "2006-01-02 15:04:05"

	// TimestampISO is for JSON/machine output (RFC3339).
	TimestampISO = time.RFC3339
)

// FormatTime formats a time pointer for display.
// Returns "-" if the time is nil or zero.
func FormatTime(t *time.Time, format string) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return t.Local().Format(format)
}

// Format specifies the output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// ParseFormat converts a string to Format, defaulting to FormatText.
func ParseFormat(s string) Format {
	switch s {
	case "json":
		return FormatJSON
	case "yaml":
		return FormatYAML
	default:
		return FormatText
	}
}

// CommandOutput is the interface for cobra.Command output methods.
type CommandOutput interface {
	OutOrStdout() io.Writer
	ErrOrStderr() io.Writer
}

// Printer provides consistent output formatting for CLI commands.
type Printer struct {
	out    io.Writer
	err    io.Writer
	format Format
}

// NewPrinter creates a Printer from a cobra.Command (or any CommandOutput).
func NewPrinter(cmd CommandOutput) *Printer {
	return &Printer{
		out:    cmd.OutOrStdout(),
		err:    cmd.ErrOrStderr(),
		format: FormatText,
	}
}

// WithFormat returns a new Printer with the specified format.
func (p *Printer) WithFormat(f Format) *Printer {
	return &Printer{out: p.out, err: p.err, format: f}
}

// IsStructured returns true if the printer is in JSON or YAML mode.
func (p *Printer) IsStructured() bool {
	return p.format == FormatJSON || p.format == FormatYAML
}

// Print prints a message to stdout.
func (p *Printer) Print(msg string) {
	_, _ = fmt.Fprintln(p.out, msg)
}

// Printf prints a formatted message to stdout.
func (p *Printer) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.out, format, args...)
}

// Errorf prints a formatted message to stderr.
func (p *Printer) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.err, format, args...)
}

// === Table output ===

// Table provides tabwriter-based table formatting.
type Table struct {
	w *tabwriter.Writer
}

// NewTable creates a new table with the given headers.
func (p *Printer) NewTable(headers ...string) *Table {
	w := tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
	if len(headers) > 0 {
		_, _ = fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	return &Table{w: w}
}

// Row adds a row to the table.
func (t *Table) Row(values ...string) {
	_, _ = fmt.Fprintln(t.w, strings.Join(values, "\t"))
}

// Flush writes the table to the output.
func (t *Table) Flush() error {
	return t.w.Flush()
}

// === Structured output (JSON/YAML) ===

// Structured writes data as JSON or YAML based on the printer's format.
func (p *Printer) Structured(data any) error {
	if p.format == FormatYAML {
		return YAMLOutput(p.out, data)
	}
	return JSONOutput(p.out, data)
}

// JSONOutput writes data as indented JSON to the writer.
func JSONOutput(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// YAMLOutput writes data as YAML to the writer.
func YAMLOutput(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(data)
}

// === High-level output helpers ===

// List prints items as either a table or structured (JSON/YAML) based on the Printer's format.
// Items should be the output type with json/yaml tags. rowFunc extracts table columns.
func List[T any](p *Printer, headers []string, items []T, rowFunc func(T) []string) error {
	if p.IsStructured() {
		return p.Structured(items)
	}
	t := p.NewTable(headers...)
	for _, item := range items {
		t.Row(rowFunc(item)...)
	}
	return t.Flush()
}
