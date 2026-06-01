package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
	"github.com/chirpz-ai/pandaprobe-cli/internal/version"
)

func renderTableStr(t *testing.T, v any) string {
	t.Helper()
	var out, errOut bytes.Buffer
	w := New(&out, &errOut, "table", true)
	require.NoError(t, w.Render(v))
	return out.String()
}

func TestSessionListTable(t *testing.T) {
	s := renderTableStr(t, models.AsList(&models.Paginated[models.SessionSummary]{
		Items: []models.SessionSummary{{SessionID: "s1", TraceCount: 3, HasError: true}},
		Total: 1, Limit: 50,
	}))
	assert.Contains(t, s, "SESSION_ID")
	assert.Contains(t, s, "s1")
	assert.Contains(t, s, "true")
}

func TestEvalRunListTable(t *testing.T) {
	s := renderTableStr(t, models.AsList(&models.Paginated[models.EvalRunResponse]{
		Items: []models.EvalRunResponse{{ID: "r1", Status: models.EvalStatusCompleted, TargetType: "trace", EvaluatedCount: 2, TotalTargets: 2}},
		Total: 1, Limit: 50,
	}))
	assert.Contains(t, s, "RUN_ID")
	assert.Contains(t, s, "r1")
	assert.Contains(t, s, "2/2")
}

func TestMetricTable(t *testing.T) {
	s := renderTableStr(t, []models.MetricSummary{{Name: "accuracy", Category: "quality", Description: "how accurate"}})
	assert.Contains(t, s, "NAME")
	assert.Contains(t, s, "accuracy")
}

func TestTraceScoreTables(t *testing.T) {
	s := renderTableStr(t, []models.TraceScoreResponse{{TraceID: "t1", Name: "acc", Value: strptr("0.9"), DataType: models.ScoreDataTypeNumeric, Source: models.ScoreSourceAutomated, Status: models.ScoreStatus("SUCCESS")}})
	assert.Contains(t, s, "TRACE_ID")
	assert.Contains(t, s, "0.9")

	s = renderTableStr(t, []models.SessionScoreResponse{{SessionID: "s1", Name: "coh", Value: strptr("1"), DataType: "NUMERIC", Source: "AUTOMATED", Status: "SUCCESS"}})
	assert.Contains(t, s, "SESSION_ID")
}

func TestTraceDetailTableWithSpans(t *testing.T) {
	s := renderTableStr(t, &models.TraceResponse{
		TraceID: "t1", Name: "n", Status: models.TraceStatusCompleted,
		Tags:  []string{"a", "b"},
		Spans: []models.SpanResponse{{SpanID: "s1", Name: "x", Kind: models.SpanKindLLM, Status: models.SpanStatusOK}},
	})
	assert.Contains(t, s, "trace_id")
	assert.Contains(t, s, "SPAN_ID")
}

func TestSessionDetailTable(t *testing.T) {
	s := renderTableStr(t, &models.SessionDetail{
		SessionSummary: models.SessionSummary{SessionID: "s1", TraceCount: 1},
		Traces:         []models.TraceResponse{{TraceID: "t1", Name: "n", Status: models.TraceStatusCompleted}},
	})
	assert.Contains(t, s, "session_id")
	assert.Contains(t, s, "TRACE_ID")
}

func TestEvalRunDetailTable(t *testing.T) {
	s := renderTableStr(t, &models.EvalRunResponse{ID: "r1", Status: models.EvalStatusFailed, TargetType: "session", MetricNames: []string{"m1"}, ErrorMessage: strptr("boom")})
	assert.Contains(t, s, "id")
	assert.Contains(t, s, "boom")
}

func TestVersionTable(t *testing.T) {
	s := renderTableStr(t, version.Info{Version: "1.2.3", OS: "darwin"})
	assert.Contains(t, s, "version")
	assert.Contains(t, s, "1.2.3")
}

func TestMapStringTable(t *testing.T) {
	s := renderTableStr(t, map[string]string{"b": "2", "a": "1"})
	// keys should be sorted: a before b
	ai := bytesIndex(s, "a")
	bi := bytesIndex(s, "b")
	assert.True(t, ai < bi)
}

func bytesIndex(s, sub string) int { return bytes.Index([]byte(s), []byte(sub)) }
