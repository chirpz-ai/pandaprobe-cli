package output

import (
	"fmt"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

// ValidationItem is one structured validation detail surfaced from a 422
// response. Defined here (rather than imported from the api package) so the
// transport layer can depend on output without creating an import cycle.
type ValidationItem struct {
	Loc  []any  `json:"loc,omitempty"`
	Msg  string `json:"msg"`
	Type string `json:"type,omitempty"`
}

// APIErrorInfo is implemented by transport errors to enrich error rendering.
type APIErrorInfo interface {
	HTTPStatus() int
	RequestID() string
	ValidationDetails() []ValidationItem
}

// errorEnvelope is the JSON error shape written to stderr.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string           `json:"code"`
	Message   string           `json:"message"`
	Status    int              `json:"status,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
	Details   []ValidationItem `json:"details,omitempty"`
}

func codeString(c exitcode.Code) string {
	switch c {
	case exitcode.Auth:
		return "auth_error"
	case exitcode.NotFound:
		return "not_found"
	case exitcode.Validation:
		return "validation_error"
	case exitcode.APIError:
		return "api_error"
	default:
		return "error"
	}
}

func buildErrorBody(err error) errorBody {
	body := errorBody{
		Code:    codeString(exitcode.From(err)),
		Message: err.Error(),
	}
	if info, ok := err.(APIErrorInfo); ok {
		body.Status = info.HTTPStatus()
		body.RequestID = info.RequestID()
		body.Details = info.ValidationDetails()
	}
	return body
}

func (w *Writer) renderErrorJSON(err error) error {
	return w.renderJSON(w.errOut, errorEnvelope{Error: buildErrorBody(err)})
}

func (w *Writer) renderErrorTable(err error) error {
	body := buildErrorBody(err)
	if _, e := fmt.Fprintf(w.errOut, "%sError:%s %s\n", w.color(colorRed), w.color(colorReset), body.Message); e != nil {
		return e
	}
	if body.Status != 0 {
		if _, e := fmt.Fprintf(w.errOut, "  status: %d\n", body.Status); e != nil {
			return e
		}
	}
	if body.RequestID != "" {
		if _, e := fmt.Fprintf(w.errOut, "  request_id: %s\n", body.RequestID); e != nil {
			return e
		}
	}
	for _, d := range body.Details {
		if _, e := fmt.Fprintf(w.errOut, "  - %v: %s\n", d.Loc, d.Msg); e != nil {
			return e
		}
	}
	if ec, ok := err.(*exitcode.Error); ok && ec.Hint != "" {
		if _, e := fmt.Fprintf(w.errOut, "  hint: %s\n", ec.Hint); e != nil {
			return e
		}
	}
	return nil
}
