package output

import (
	"encoding/json"
	"io"
)

func (w *Writer) renderJSON(dst io.Writer, v any) error {
	enc := json.NewEncoder(dst)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
