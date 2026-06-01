package api

import (
	"context"
	"net/url"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// --- Metrics ---

// TraceMetrics calls GET /evaluations/trace-metrics.
func (c *Client) TraceMetrics(ctx context.Context) (*[]models.MetricSummary, error) {
	return doDecode[[]models.MetricSummary](ctx, c, "GET", "/evaluations/trace-metrics", nil, nil)
}

// SessionMetrics calls GET /evaluations/session-metrics.
func (c *Client) SessionMetrics(ctx context.Context) (*[]models.MetricSummary, error) {
	return doDecode[[]models.MetricSummary](ctx, c, "GET", "/evaluations/session-metrics", nil, nil)
}

// --- Runs (create) ---

// CreateTraceRun calls POST /evaluations/trace-runs.
func (c *Client) CreateTraceRun(ctx context.Context, body models.CreateEvalRunRequest) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "POST", "/evaluations/trace-runs", nil, body)
}

// BatchTraceRun calls POST /evaluations/trace-runs/batch.
func (c *Client) BatchTraceRun(ctx context.Context, body models.CreateBatchEvalRunRequest) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "POST", "/evaluations/trace-runs/batch", nil, body)
}

// CreateSessionRun calls POST /evaluations/session-runs.
func (c *Client) CreateSessionRun(ctx context.Context, body models.CreateSessionEvalRunRequest) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "POST", "/evaluations/session-runs", nil, body)
}

// BatchSessionRun calls POST /evaluations/session-runs/batch.
func (c *Client) BatchSessionRun(ctx context.Context, body models.CreateBatchSessionEvalRunRequest) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "POST", "/evaluations/session-runs/batch", nil, body)
}

// --- Runs (read) ---

// ListTraceRuns calls GET /evaluations/trace-runs.
func (c *Client) ListTraceRuns(ctx context.Context, q url.Values) (*models.Paginated[models.EvalRunResponse], error) {
	return doDecode[models.Paginated[models.EvalRunResponse]](ctx, c, "GET", "/evaluations/trace-runs", q, nil)
}

// ListSessionRuns calls GET /evaluations/session-runs.
func (c *Client) ListSessionRuns(ctx context.Context, q url.Values) (*models.Paginated[models.EvalRunResponse], error) {
	return doDecode[models.Paginated[models.EvalRunResponse]](ctx, c, "GET", "/evaluations/session-runs", q, nil)
}

// GetTraceRun calls GET /evaluations/trace-runs/{run_id}.
func (c *Client) GetTraceRun(ctx context.Context, runID string) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "GET", "/evaluations/trace-runs/"+url.PathEscape(runID), nil, nil)
}

// GetSessionRun calls GET /evaluations/session-runs/{run_id}.
func (c *Client) GetSessionRun(ctx context.Context, runID string) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "GET", "/evaluations/session-runs/"+url.PathEscape(runID), nil, nil)
}

// TraceRunScores calls GET /evaluations/trace-runs/{run_id}/scores.
func (c *Client) TraceRunScores(ctx context.Context, runID string) (*[]models.TraceScoreResponse, error) {
	return doDecode[[]models.TraceScoreResponse](ctx, c, "GET", "/evaluations/trace-runs/"+url.PathEscape(runID)+"/scores", nil, nil)
}

// SessionRunScores calls GET /evaluations/session-runs/{run_id}/scores.
func (c *Client) SessionRunScores(ctx context.Context, runID string) (*[]models.SessionScoreResponse, error) {
	return doDecode[[]models.SessionScoreResponse](ctx, c, "GET", "/evaluations/session-runs/"+url.PathEscape(runID)+"/scores", nil, nil)
}

// --- Scores ---

// ListTraceScores calls GET /evaluations/trace-scores.
func (c *Client) ListTraceScores(ctx context.Context, q url.Values) (*models.Paginated[models.TraceScoreResponse], error) {
	return doDecode[models.Paginated[models.TraceScoreResponse]](ctx, c, "GET", "/evaluations/trace-scores", q, nil)
}

// ListSessionScores calls GET /evaluations/session-scores.
func (c *Client) ListSessionScores(ctx context.Context, q url.Values) (*models.Paginated[models.SessionScoreResponse], error) {
	return doDecode[models.Paginated[models.SessionScoreResponse]](ctx, c, "GET", "/evaluations/session-scores", q, nil)
}

// GetTraceScores calls GET /evaluations/trace-scores/{trace_id}.
func (c *Client) GetTraceScores(ctx context.Context, traceID string) (*[]models.TraceScoreResponse, error) {
	return doDecode[[]models.TraceScoreResponse](ctx, c, "GET", "/evaluations/trace-scores/"+url.PathEscape(traceID), nil, nil)
}

// GetSessionScores calls GET /evaluations/session-scores/{session_id}.
func (c *Client) GetSessionScores(ctx context.Context, sessionID string) (*[]models.SessionScoreResponse, error) {
	return doDecode[[]models.SessionScoreResponse](ctx, c, "GET", "/evaluations/session-scores/"+url.PathEscape(sessionID), nil, nil)
}

// SubmitTraceScore calls POST /evaluations/trace-scores.
func (c *Client) SubmitTraceScore(ctx context.Context, body models.CreateTraceScoreRequest) (*models.TraceScoreResponse, error) {
	return doDecode[models.TraceScoreResponse](ctx, c, "POST", "/evaluations/trace-scores", nil, body)
}
