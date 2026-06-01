package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/config"
	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func testClient(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg := &config.Config{
		Endpoint:    srv.URL,
		APIKey:      "sk_pp_secret123",
		ProjectName: "proj",
		Timeout:     5 * time.Second,
	}
	allOpts := append([]Option{WithRequestID(func() string { return "fixed-req-id" })}, opts...)
	return New(cfg, nil, allOpts...), srv
}

func TestListTracesSendsHeadersAndQuery(t *testing.T) {
	var gotReq *http.Request
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":50,"offset":0}`))
	})

	q := map[string][]string{"status": {"COMPLETED"}, "limit": {"50"}}
	_, err := c.ListTraces(context.Background(), q)
	require.NoError(t, err)

	assert.Equal(t, "sk_pp_secret123", gotReq.Header.Get("X-API-Key"))
	assert.Equal(t, "proj", gotReq.Header.Get("X-Project-Name"))
	assert.Equal(t, "fixed-req-id", gotReq.Header.Get("X-Request-ID"))
	assert.Contains(t, gotReq.Header.Get("User-Agent"), "pandaprobe-cli/")
	assert.Equal(t, "/traces", gotReq.URL.Path)
	assert.Equal(t, "COMPLETED", gotReq.URL.Query().Get("status"))
}

func TestGetTraceDecodes(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/traces/abc", r.URL.Path)
		_, _ = w.Write([]byte(`{"trace_id":"abc","name":"x","status":"COMPLETED","spans":[{"span_id":"s1","trace_id":"abc","name":"n","kind":"LLM","status":"OK"}]}`))
	})
	tr, err := c.GetTrace(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "abc", tr.TraceID)
	require.Len(t, tr.Spans, 1)
	assert.Equal(t, models.SpanKindLLM, tr.Spans[0].Kind)
}

func TestPostSendsContentTypeAndBody(t *testing.T) {
	var gotBody []byte
	var gotCT string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = readAll(r)
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"sc1","trace_id":"t1","name":"acc","status":"SUCCESS","source":"PROGRAMMATIC","data_type":"NUMERIC","created_at":"","updated_at":"","project_id":"p"}`))
	})
	res, err := c.SubmitTraceScore(context.Background(), models.CreateTraceScoreRequest{TraceID: "t1", Name: "acc", Value: "0.9"})
	require.NoError(t, err)
	assert.Equal(t, "sc1", res.ID)
	assert.Equal(t, "application/json", gotCT)
	assert.Contains(t, string(gotBody), `"trace_id":"t1"`)
}

func TestErrorStatusMapping(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   int // exitcode.Code
	}{
		{401, `{"detail":"bad key"}`, 2},
		{403, `{"detail":"forbidden"}`, 2},
		{404, `{"detail":"missing"}`, 3},
		{400, `{"detail":"bad"}`, 4},
		{422, `{"detail":[{"loc":["query","status"],"msg":"invalid","type":"enum"}]}`, 4},
		{500, `{"detail":"boom"}`, 5},
		{503, `oops not json`, 5},
	}
	for _, tt := range tests {
		c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
			_, _ = w.Write([]byte(tt.body))
		})
		_, err := c.GetTrace(context.Background(), "x")
		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok, "want *APIError for status %d", tt.status)
		assert.Equal(t, tt.status, apiErr.HTTPStatus())
		assert.Equal(t, tt.want, int(apiErr.ExitCode()))
		assert.Equal(t, "fixed-req-id", apiErr.RequestID())
	}
}

func TestValidationDetailsParsed(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"detail":[{"loc":["query","status"],"msg":"invalid","type":"enum"}]}`))
	})
	_, err := c.GetTrace(context.Background(), "x")
	apiErr := err.(*APIError)
	require.Len(t, apiErr.Validation, 1)
	assert.Equal(t, "invalid", apiErr.Validation[0].Msg)
	assert.Len(t, apiErr.ValidationDetails(), 1)
}

func TestDebugMasksAPIKey(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{Endpoint: "", APIKey: "sk_pp_supersecret", ProjectName: "p", Timeout: time.Second, Debug: true}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":50,"offset":0}`))
	}))
	defer srv.Close()
	cfg.Endpoint = srv.URL
	c := New(cfg, &buf, WithRequestID(func() string { return "rid" }))

	_, err := c.ListTraces(context.Background(), nil)
	require.NoError(t, err)
	logs := buf.String()
	assert.Contains(t, logs, "[debug] req GET")
	assert.NotContains(t, logs, "supersecret")
	assert.Contains(t, logs, "sk_pp_****cret")
}

func TestContextCancellation(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.ListTraces(ctx, nil)
	require.Error(t, err)
	_, isAPI := err.(*APIError)
	assert.False(t, isAPI, "transport error should not be an APIError")
}

func readAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	return buf.Bytes(), err
}
