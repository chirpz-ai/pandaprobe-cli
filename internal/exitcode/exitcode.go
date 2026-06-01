// Package exitcode defines the process exit codes used by the CLI and a typed
// error that carries one. Exit codes are part of the CLI's contract: agents
// parse them to decide how to react.
package exitcode

import "fmt"

// Code is a process exit status.
type Code int

// Exit codes. These are a stable part of the CLI contract.
const (
	OK         Code = 0 // success
	General    Code = 1 // unexpected/general failure (network, decode, etc.)
	Auth       Code = 2 // authentication/authorization failure (401, 403)
	NotFound   Code = 3 // resource not found (404)
	Validation Code = 4 // client-side validation or 400/422 from the server
	APIError   Code = 5 // other server-side error (other 4xx, 5xx)
)

// Error is an error that carries an explicit exit code. Client-side validation
// failures and configuration problems use this so the run wrapper can both
// render them and set the right process exit status.
type Error struct {
	Code    Code
	Message string
	// Hint is optional supplementary guidance rendered in human mode.
	Hint string
}

func (e *Error) Error() string { return e.Message }

// New builds a *Error with the given code and formatted message.
func New(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Coder is implemented by errors that know their own exit code (e.g. *Error and
// the API error type).
type Coder interface {
	ExitCode() Code
}

// ExitCode implements Coder.
func (e *Error) ExitCode() Code { return e.Code }

// From resolves the process exit code for an error returned from a command.
// nil -> OK; anything implementing Coder uses its code; everything else is a
// General failure.
func From(err error) Code {
	if err == nil {
		return OK
	}
	if c, ok := err.(Coder); ok {
		return c.ExitCode()
	}
	return General
}
