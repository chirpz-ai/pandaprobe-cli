package cmd

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

// addPaginationFlags registers --limit and --offset. The API caps limit at
// maxLimit; an unset limit (0) lets the server apply its default.
func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 0, "Max results to return (1-200; 0 uses server default)")
	cmd.Flags().Int("offset", 0, "Number of results to skip")
}

// validateEnum returns an error if val is non-empty and not in allowed.
func validateEnum(flag, val string, allowed []string) error {
	if val == "" {
		return nil
	}
	for _, a := range allowed {
		if a == val {
			return nil
		}
	}
	return exitcode.New(exitcode.Validation,
		"invalid --%s %q: must be one of %s", flag, val, strings.Join(allowed, ", "))
}

// parseDate accepts RFC3339 or YYYY-MM-DD and returns a normalized RFC3339
// string (UTC midnight for date-only input).
func parseDate(flag, val string) (string, error) {
	if val == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t.Format(time.RFC3339), nil
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", exitcode.New(exitcode.Validation,
		"invalid --%s %q: use RFC3339 (2025-01-15T10:30:00Z) or YYYY-MM-DD", flag, val)
}

// query accumulates URL query parameters from flags, applying validation and
// only including parameters whose flags the user actually set. Errors are
// captured and surfaced by build() so callers can chain fluently.
type query struct {
	v   url.Values
	cmd *cobra.Command
	err error
}

func newQuery(cmd *cobra.Command) *query {
	return &query{v: url.Values{}, cmd: cmd}
}

func (q *query) changed(flag string) bool {
	f := q.cmd.Flags().Lookup(flag)
	return f != nil && f.Changed
}

// str sets param to the string flag's value when the flag was set.
func (q *query) str(param, flag string) *query {
	if q.err != nil || !q.changed(flag) {
		return q
	}
	val, _ := q.cmd.Flags().GetString(flag)
	q.v.Set(param, val)
	return q
}

// enum is like str but validates against allowed values first.
func (q *query) enum(param, flag string, allowed []string) *query {
	if q.err != nil || !q.changed(flag) {
		return q
	}
	val, _ := q.cmd.Flags().GetString(flag)
	if err := validateEnum(flag, val, allowed); err != nil {
		q.err = err
		return q
	}
	q.v.Set(param, val)
	return q
}

// date parses and normalizes a date flag.
func (q *query) date(param, flag string) *query {
	if q.err != nil || !q.changed(flag) {
		return q
	}
	val, _ := q.cmd.Flags().GetString(flag)
	norm, err := parseDate(flag, val)
	if err != nil {
		q.err = err
		return q
	}
	q.v.Set(param, norm)
	return q
}

// strs adds one query param per element of a string-slice flag.
func (q *query) strs(param, flag string) *query {
	if q.err != nil || !q.changed(flag) {
		return q
	}
	vals, _ := q.cmd.Flags().GetStringSlice(flag)
	for _, val := range vals {
		q.v.Add(param, val)
	}
	return q
}

// boolean sets param to "true"/"false" when a bool flag was set.
func (q *query) boolean(param, flag string) *query {
	if q.err != nil || !q.changed(flag) {
		return q
	}
	val, _ := q.cmd.Flags().GetBool(flag)
	q.v.Set(param, strconv.FormatBool(val))
	return q
}

// pagination validates and sets limit/offset. maxLimit bounds the limit flag.
func (q *query) pagination(maxLimit int) *query {
	if q.err != nil {
		return q
	}
	if q.changed("limit") {
		limit, _ := q.cmd.Flags().GetInt("limit")
		if limit < 1 || limit > maxLimit {
			q.err = exitcode.New(exitcode.Validation, "invalid --limit %d: must be between 1 and %d", limit, maxLimit)
			return q
		}
		q.v.Set("limit", strconv.Itoa(limit))
	}
	if q.changed("offset") {
		offset, _ := q.cmd.Flags().GetInt("offset")
		if offset < 0 {
			q.err = exitcode.New(exitcode.Validation, "invalid --offset %d: must be >= 0", offset)
			return q
		}
		q.v.Set("offset", strconv.Itoa(offset))
	}
	return q
}

func (q *query) build() (url.Values, error) {
	return q.v, q.err
}

// strPtrFlag returns a pointer to the flag's value, or nil if it was not set.
func strPtrFlag(cmd *cobra.Command, flag string) *string {
	f := cmd.Flags().Lookup(flag)
	if f == nil || !f.Changed {
		return nil
	}
	val, _ := cmd.Flags().GetString(flag)
	return &val
}
