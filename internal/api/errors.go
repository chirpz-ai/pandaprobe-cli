package api

import (
	"encoding/json"
	"fmt"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/output"
)

// ValidationDetail is one item from a 422 HTTPValidationError response.
type ValidationDetail struct {
	Loc   []any  `json:"loc"`
	Msg   string `json:"msg"`
	Type  string `json:"type"`
	Input any    `json:"input,omitempty"`
}

// APIError is a non-2xx response from the backend. It implements
// exitcode.Coder and output.APIErrorInfo so the run wrapper and the output
// layer can render it consistently.
type APIError struct {
	Status     int
	Detail     string
	Validation []ValidationDetail
	ReqID      string
	Raw        []byte
}

func (e *APIError) Error() string {
	if len(e.Validation) > 0 {
		first := e.Validation[0]
		return fmt.Sprintf("validation error (HTTP %d): %v: %s", e.Status, first.Loc, first.Msg)
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s (HTTP %d)", e.Detail, e.Status)
	}
	return fmt.Sprintf("request failed with HTTP %d", e.Status)
}

// ExitCode maps the HTTP status to the CLI's exit-code contract.
func (e *APIError) ExitCode() exitcode.Code {
	switch e.Status {
	case 401, 403:
		return exitcode.Auth
	case 404:
		return exitcode.NotFound
	case 400, 422:
		return exitcode.Validation
	default:
		return exitcode.APIError
	}
}

// HTTPStatus implements output.APIErrorInfo.
func (e *APIError) HTTPStatus() int { return e.Status }

// RequestID implements output.APIErrorInfo.
func (e *APIError) RequestID() string { return e.ReqID }

// ValidationDetails implements output.APIErrorInfo.
func (e *APIError) ValidationDetails() []output.ValidationItem {
	if len(e.Validation) == 0 {
		return nil
	}
	items := make([]output.ValidationItem, 0, len(e.Validation))
	for _, d := range e.Validation {
		items = append(items, output.ValidationItem{Loc: d.Loc, Msg: d.Msg, Type: d.Type})
	}
	return items
}

// Parse turns a non-2xx response body into an *APIError. It accepts both the
// 422 shape ({"detail":[{...}]}) and the generic shape ({"detail":"..."}),
// falling back to the raw body when neither matches.
func Parse(status int, body []byte, reqID string) *APIError {
	e := &APIError{Status: status, ReqID: reqID, Raw: body}

	var envelope struct {
		Detail json.RawMessage `json:"detail"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Detail) > 0 {
		// Try the validation array shape first.
		var items []ValidationDetail
		if err := json.Unmarshal(envelope.Detail, &items); err == nil && len(items) > 0 {
			e.Validation = items
			return e
		}
		// Then the plain string shape.
		var s string
		if err := json.Unmarshal(envelope.Detail, &s); err == nil {
			e.Detail = s
			return e
		}
	}

	// No recognizable envelope: surface the trimmed raw body if present.
	if len(body) > 0 {
		e.Detail = string(body)
	}
	return e
}
