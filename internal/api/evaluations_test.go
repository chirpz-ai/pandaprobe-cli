package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func TestSessionsAndMetrics(t *testing.T) {
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions":
			_, _ = w.Write([]byte(`{"items":[{"session_id":"s1","trace_count":2,"first_trace_at":"","has_error":false,"tags":[]}],"total":1,"limit":50,"offset":0}`))
		case "/sessions/s1":
			_, _ = w.Write([]byte(`{"session_id":"s1","trace_count":1,"first_trace_at":"","has_error":false,"tags":[],"traces":[]}`))
		case "/evaluations/trace-metrics":
			_, _ = w.Write([]byte(`[{"name":"accuracy","description":"d","category":"quality"}]`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	ctx := context.Background()

	sessions, err := c.ListSessions(ctx, nil)
	require.NoError(t, err)
	require.Len(t, sessions.Items, 1)
	assert.Equal(t, "s1", sessions.Items[0].SessionID)

	detail, err := c.GetSession(ctx, "s1", nil)
	require.NoError(t, err)
	assert.Equal(t, "s1", detail.SessionID)
	assert.NotNil(t, detail.Traces)

	metrics, err := c.TraceMetrics(ctx)
	require.NoError(t, err)
	require.Len(t, *metrics, 1)
	assert.Equal(t, "accuracy", (*metrics)[0].Name)
}

func TestEvalRunsLifecycle(t *testing.T) {
	var lastPath, lastMethod string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		lastPath, lastMethod = r.URL.Path, r.Method
		switch {
		case r.URL.Path == "/evaluations/trace-runs" && r.Method == "POST":
			w.WriteHeader(202)
			_, _ = w.Write([]byte(`{"id":"run1","status":"PENDING","metric_names":["accuracy"],"total_targets":0,"evaluated_count":0,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}`))
		case r.URL.Path == "/evaluations/trace-runs" && r.Method == "GET":
			_, _ = w.Write([]byte(`{"items":[{"id":"run1","status":"COMPLETED","metric_names":[],"total_targets":3,"evaluated_count":3,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}],"total":1,"limit":50,"offset":0}`))
		case r.URL.Path == "/evaluations/trace-runs/run1":
			_, _ = w.Write([]byte(`{"id":"run1","status":"COMPLETED","metric_names":[],"total_targets":3,"evaluated_count":3,"failed_count":0,"created_at":"","project_id":"p","target_type":"trace","filters":{},"sampling_rate":1}`))
		case r.URL.Path == "/evaluations/trace-runs/run1/scores":
			_, _ = w.Write([]byte(`[{"id":"sc1","trace_id":"t1","name":"accuracy","status":"SUCCESS","source":"AUTOMATED","data_type":"NUMERIC","created_at":"","updated_at":"","project_id":"p","metadata":{}}]`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	ctx := context.Background()

	created, err := c.CreateTraceRun(ctx, models.CreateEvalRunRequest{Metrics: []string{"accuracy"}})
	require.NoError(t, err)
	assert.Equal(t, "run1", created.ID)
	assert.Equal(t, models.EvalStatusPending, created.Status)

	list, err := c.ListTraceRuns(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Items, 1)

	got, err := c.GetTraceRun(ctx, "run1")
	require.NoError(t, err)
	assert.Equal(t, models.EvalStatusCompleted, got.Status)

	scores, err := c.TraceRunScores(ctx, "run1")
	require.NoError(t, err)
	require.Len(t, *scores, 1)
	assert.Equal(t, "/evaluations/trace-runs/run1/scores", lastPath)
	assert.Equal(t, "GET", lastMethod)
}
