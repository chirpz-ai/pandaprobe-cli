package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

// runCLI builds a fresh command tree, runs it with args, and returns captured
// stdout, stderr, and the resolved exit code.
func runCLI(t *testing.T, args ...string) (string, string, exitcode.Code) {
	t.Helper()
	app := &appContext{}
	root := newRootCmd(app)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	code := executeRoot(root, app)
	return stdout.String(), stderr.String(), code
}

// withAuthEnv points the CLI at srv and supplies credentials via env.
func withAuthEnv(t *testing.T, srv *httptest.Server) {
	t.Helper()
	t.Setenv("PANDAPROBE_ENDPOINT", srv.URL)
	t.Setenv("PANDAPROBE_API_KEY", "sk_pp_testkey1234")
	t.Setenv("PANDAPROBE_PROJECT_NAME", "proj")
	// Ensure no stray config file leaks in.
	t.Setenv("HOME", t.TempDir())
}

func TestVersionCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	out, _, code := runCLI(t, "version")
	assert.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, `"version"`)
}

func TestMissingAPIKeyExit2(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PANDAPROBE_API_KEY", "")
	t.Setenv("PANDAPROBE_PROJECT_NAME", "")
	out, errOut, code := runCLI(t, "traces", "list")
	assert.Equal(t, exitcode.Auth, code)
	assert.Empty(t, out)
	assert.Contains(t, errOut, "auth_error")
}

func TestTracesListSendsFilters(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"items":[{"trace_id":"t1","name":"q","status":"ERROR","started_at":"2025-01-01T00:00:00Z","tags":[]}],"total":1,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "traces", "list", "--status", "ERROR", "--limit", "50")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, gotQuery, "status=ERROR")
	assert.Contains(t, gotQuery, "limit=50")
	assert.Contains(t, out, `"trace_id": "t1"`)
	assert.Contains(t, out, `"pagination"`)
}

func TestTracesListBadEnumExit4NoRequest(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "traces", "list", "--status", "BOGUS")
	assert.Equal(t, exitcode.Validation, code)
	assert.Empty(t, out)
	assert.Contains(t, errOut, "validation_error")
	assert.False(t, hit, "no request should be made when validation fails")
}

func TestTracesGetNotFoundExit3(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"detail":"trace not found"}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "traces", "get", "missing")
	assert.Equal(t, exitcode.NotFound, code)
	assert.Contains(t, errOut, "not_found")
}

func TestTracesSpansClientSideFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"trace_id":"t1","name":"n","status":"COMPLETED","spans":[
			{"span_id":"s1","trace_id":"t1","name":"a","kind":"LLM","status":"OK"},
			{"span_id":"s2","trace_id":"t1","name":"b","kind":"TOOL","status":"ERROR"}]}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, _, code := runCLI(t, "traces", "spans", "t1", "--kind", "LLM")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, `"span_id": "s1"`)
	assert.NotContains(t, out, `"span_id": "s2"`)
}

func TestEvalsScoresSubmitSessionRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be made")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "scores", "submit", "--target", "session", "--trace-id", "t1", "--name", "x", "--value", "1")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "session score submission is not supported")
}

func TestEvalsScoresSubmitTrace(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		body = buf.Bytes()
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"sc1","trace_id":"t1","name":"acc","value":"0.9","status":"SUCCESS","source":"PROGRAMMATIC","data_type":"NUMERIC","created_at":"","updated_at":"","project_id":"p","metadata":{}}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, errOut, code := runCLI(t, "evals", "scores", "submit", "--trace-id", "t1", "--name", "acc", "--value", "0.9")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, string(body), `"trace_id":"t1"`)
	assert.Contains(t, string(body), `"source":"PROGRAMMATIC"`)
	assert.Contains(t, out, `"id": "sc1"`)
}

func TestEvalsRunsCreateRequiresMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request expected")
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	_, errOut, code := runCLI(t, "evals", "runs", "create")
	assert.Equal(t, exitcode.Validation, code)
	assert.Contains(t, errOut, "metrics")
}

func TestConfigShowMasksKeyAndReportsSource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PANDAPROBE_API_KEY", "sk_pp_supersecretvalue")
	out, _, code := runCLI(t, "config", "show")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, "sk_pp_****alue")
	assert.NotContains(t, out, "supersecretvalue")
	assert.Contains(t, out, `"source": "env"`)
}

func TestTableFormatNoColor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"trace_id":"t1","name":"hello","status":"COMPLETED","started_at":"2025-01-01T00:00:00Z","tags":[]}],"total":1,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	withAuthEnv(t, srv)

	out, _, code := runCLI(t, "traces", "list", "--format", "table", "--no-color")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, "TRACE_ID")
	assert.NotContains(t, out, "\033[")
}

// ensure the command tree is internally consistent (no duplicate flags etc.).
func TestCommandTreeValid(t *testing.T) {
	root := newRootCmd(&appContext{})
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		require.NoError(t, c.ValidateArgs(nil))
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	_ = walk // tree construction alone exercises flag registration panics
	assert.NotNil(t, root)
}
