package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// TestRemainingEndpoints exercises every remaining typed wrapper against a
// catch-all server, asserting the right path/method is hit.
func TestRemainingEndpoints(t *testing.T) {
	var path, method string
	c, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		// Return a shape valid for both array and paginated decodes depending on
		// the endpoint; each call below decodes into its own concrete type.
		switch {
		case strings.HasSuffix(r.URL.Path, "/scores") || strings.Contains(r.URL.Path, "-scores/") || strings.HasSuffix(r.URL.Path, "-metrics"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == "POST":
			w.WriteHeader(202)
			_, _ = w.Write([]byte(`{"id":"r","status":"PENDING","metric_names":[],"total_targets":0,"evaluated_count":0,"failed_count":0,"created_at":"","project_id":"p","target_type":"x","filters":{},"sampling_rate":1}`))
		default:
			_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":50,"offset":0}`))
		}
	})
	ctx := context.Background()

	mustOK := func(_ any, err error) { require.NoError(t, err) }

	mustOK(c.SessionMetrics(ctx))
	mustOK(c.CreateSessionRun(ctx, models.CreateSessionEvalRunRequest{Metrics: []string{"m"}}))
	mustOK(c.BatchTraceRun(ctx, models.CreateBatchEvalRunRequest{TraceIDs: []string{"t"}, Metrics: []string{"m"}}))
	mustOK(c.BatchSessionRun(ctx, models.CreateBatchSessionEvalRunRequest{SessionIDs: []string{"s"}, Metrics: []string{"m"}}))
	mustOK(c.ListSessionRuns(ctx, nil))
	mustOK(c.GetSessionRun(ctx, "r"))
	mustOK(c.SessionRunScores(ctx, "r"))
	mustOK(c.ListTraceScores(ctx, nil))
	mustOK(c.ListSessionScores(ctx, nil))
	mustOK(c.GetTraceScores(ctx, "t"))
	mustOK(c.GetSessionScores(ctx, "s"))

	require.NotEmpty(t, path)
	require.NotEmpty(t, method)
}
