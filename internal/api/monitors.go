package api

import (
	"context"
	"net/url"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// --- Monitors (create/read) ---

// CreateMonitor calls POST /evaluations/monitors.
func (c *Client) CreateMonitor(ctx context.Context, body models.CreateMonitorRequest) (*models.MonitorResponse, error) {
	return doDecode[models.MonitorResponse](ctx, c, "POST", "/evaluations/monitors", nil, body)
}

// ListMonitors calls GET /evaluations/monitors.
func (c *Client) ListMonitors(ctx context.Context, q url.Values) (*models.Paginated[models.MonitorResponse], error) {
	return doDecode[models.Paginated[models.MonitorResponse]](ctx, c, "GET", "/evaluations/monitors", q, nil)
}

// GetMonitor calls GET /evaluations/monitors/{monitor_id}.
func (c *Client) GetMonitor(ctx context.Context, id string) (*models.MonitorResponse, error) {
	return doDecode[models.MonitorResponse](ctx, c, "GET", "/evaluations/monitors/"+url.PathEscape(id), nil, nil)
}

// --- Monitors (mutate) ---

// UpdateMonitor calls PATCH /evaluations/monitors/{monitor_id}.
func (c *Client) UpdateMonitor(ctx context.Context, id string, body models.UpdateMonitorRequest) (*models.MonitorResponse, error) {
	return doDecode[models.MonitorResponse](ctx, c, "PATCH", "/evaluations/monitors/"+url.PathEscape(id), nil, body)
}

// DeleteMonitor calls DELETE /evaluations/monitors/{monitor_id}. The endpoint
// returns 204 with no body.
func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	_, err := c.Do(ctx, "DELETE", "/evaluations/monitors/"+url.PathEscape(id), nil, nil)
	return err
}

// PauseMonitor calls POST /evaluations/monitors/{monitor_id}/pause.
func (c *Client) PauseMonitor(ctx context.Context, id string) (*models.MonitorResponse, error) {
	return doDecode[models.MonitorResponse](ctx, c, "POST", "/evaluations/monitors/"+url.PathEscape(id)+"/pause", nil, nil)
}

// ResumeMonitor calls POST /evaluations/monitors/{monitor_id}/resume.
func (c *Client) ResumeMonitor(ctx context.Context, id string) (*models.MonitorResponse, error) {
	return doDecode[models.MonitorResponse](ctx, c, "POST", "/evaluations/monitors/"+url.PathEscape(id)+"/resume", nil, nil)
}

// MonitorRuns calls GET /evaluations/monitors/{monitor_id}/runs.
func (c *Client) MonitorRuns(ctx context.Context, id string, q url.Values) (*models.Paginated[models.EvalRunResponse], error) {
	return doDecode[models.Paginated[models.EvalRunResponse]](ctx, c, "GET", "/evaluations/monitors/"+url.PathEscape(id)+"/runs", q, nil)
}

// TriggerMonitor calls POST /evaluations/monitors/{monitor_id}/trigger. The
// endpoint returns 202 with the created evaluation run.
func (c *Client) TriggerMonitor(ctx context.Context, id string) (*models.EvalRunResponse, error) {
	return doDecode[models.EvalRunResponse](ctx, c, "POST", "/evaluations/monitors/"+url.PathEscape(id)+"/trigger", nil, nil)
}
