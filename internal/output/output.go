// Package output renders command results. JSON is the default, machine-first
// format (data to stdout, errors to stdout-or-stderr as JSON). A human-oriented
// table format is opt-in. A single Render(any) entrypoint keeps the surface
// small and avoids a per-type interface.
package output

import (
	"io"
	"os"
)

// Format is an output format.
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
)

// Writer renders values in the configured format.
type Writer struct {
	out     io.Writer
	errOut  io.Writer
	format  Format
	noColor bool
}

// New builds a Writer. format defaults to JSON for any unrecognized value.
func New(out, errOut io.Writer, format string, noColor bool) *Writer {
	f := FormatJSON
	if Format(format) == FormatTable {
		f = FormatTable
	}
	// Honor the NO_COLOR convention and non-tty stdout in addition to the flag.
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	return &Writer{out: out, errOut: errOut, format: f, noColor: noColor}
}

// Format reports the writer's active format.
func (w *Writer) Format() Format { return w.format }

// Render writes a successful result to stdout.
func (w *Writer) Render(v any) error {
	if w.format == FormatTable {
		return w.renderTable(v)
	}
	return w.renderJSON(w.out, v)
}

// RenderError writes an error to stderr in the active format.
func (w *Writer) RenderError(err error) error {
	if w.format == FormatTable {
		return w.renderErrorTable(err)
	}
	return w.renderErrorJSON(err)
}

// MaskSecret redacts a secret for display, keeping a short prefix and the last
// four characters so it remains identifiable without being usable.
func MaskSecret(s string) string {
	if s == "" {
		return ""
	}
	const tailLen = 4
	if len(s) <= tailLen+2 {
		return "****"
	}
	// Keep up to the first underscore-delimited prefix (e.g. "sk_pp") plus tail.
	head := ""
	if i := indexNthSep(s, 2); i > 0 && i < len(s)-tailLen {
		head = s[:i+1]
	}
	return head + "****" + s[len(s)-tailLen:]
}

// indexNthSep returns the index of the nth '_' in s, or -1.
func indexNthSep(s string, n int) int {
	count := 0
	for i, r := range s {
		if r == '_' {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}

func truncate(s string, n int) string {
	if n <= 1 || len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
