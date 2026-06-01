package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func strptr(s string) *string { return &s }

func TestRenderJSONList(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "json", false)

	list := models.AsList(&models.Paginated[models.TraceListItem]{
		Items: []models.TraceListItem{{TraceID: "t1", Name: "n", Status: models.TraceStatusCompleted}},
		Total: 1, Limit: 50, Offset: 0,
	})
	require.NoError(t, w.Render(list))

	s := out.String()
	assert.Contains(t, s, `"items"`)
	assert.Contains(t, s, `"pagination": {`)
	assert.Contains(t, s, `"total": 1`)
	assert.Contains(t, s, `"trace_id": "t1"`)
	assert.Empty(t, errOut.String())
}

func TestRenderErrorJSONToStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "json", false)

	err := exitcode.New(exitcode.Validation, "bad flag")
	require.NoError(t, w.RenderError(err))

	assert.Empty(t, out.String())
	s := errOut.String()
	assert.Contains(t, s, `"error"`)
	assert.Contains(t, s, `"code": "validation_error"`)
	assert.Contains(t, s, `"message": "bad flag"`)
}

// fakeAPIErr exercises the APIErrorInfo path without importing the api package.
type fakeAPIErr struct{}

func (fakeAPIErr) Error() string           { return "validation failed" }
func (fakeAPIErr) ExitCode() exitcode.Code { return exitcode.Validation }
func (fakeAPIErr) HTTPStatus() int         { return 422 }
func (fakeAPIErr) RequestID() string       { return "req-123" }
func (fakeAPIErr) ValidationDetails() []ValidationItem {
	return []ValidationItem{{Loc: []any{"query", "status"}, Msg: "invalid", Type: "enum"}}
}

func TestRenderErrorWithDetails(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "json", false)
	require.NoError(t, w.RenderError(fakeAPIErr{}))
	s := errOut.String()
	assert.Contains(t, s, `"status": 422`)
	assert.Contains(t, s, `"request_id": "req-123"`)
	assert.Contains(t, s, `"details"`)
	assert.Contains(t, s, `"msg": "invalid"`)
}

func TestMaskSecret(t *testing.T) {
	assert.Equal(t, "", MaskSecret(""))
	assert.Equal(t, "****", MaskSecret("abc"))
	assert.Equal(t, "sk_pp_****3456", MaskSecret("sk_pp_abc123456"))
	// No second separator: still masked, just no head.
	assert.Equal(t, "****cdef", MaskSecret("abcdefcdef"))
}

func TestTableNoColorHasNoANSI(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "table", true)
	list := models.AsList(&models.Paginated[models.TraceListItem]{
		Items: []models.TraceListItem{{TraceID: "t1", Name: "hello", Status: models.TraceStatusError}},
		Total: 1, Limit: 50,
	})
	require.NoError(t, w.Render(list))
	s := out.String()
	assert.NotContains(t, s, "\033[")
	assert.Contains(t, s, "TRACE_ID")
	assert.Contains(t, s, "t1")
	assert.Contains(t, s, "(total 1, limit 50, offset 0)")
}

func TestTableColorOnStatus(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "table", false)
	require.NoError(t, w.Render([]models.SpanResponse{{SpanID: "s1", Name: "n", Kind: models.SpanKindLLM, Status: models.SpanStatusError, Model: strptr("gpt-4")}}))
	assert.Contains(t, out.String(), colorRed)
}

func TestUnknownTypeFallsBackToJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "table", true)
	require.NoError(t, w.Render(struct {
		Foo string `json:"foo"`
	}{Foo: "bar"}))
	assert.Contains(t, out.String(), `"foo": "bar"`)
}

func TestTableErrorToStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "table", true)
	require.NoError(t, w.RenderError(fakeAPIErr{}))
	s := errOut.String()
	assert.True(t, strings.HasPrefix(s, "Error:"))
	assert.Contains(t, s, "status: 422")
	assert.Contains(t, s, "request_id: req-123")
}
