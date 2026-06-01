package api

import (
	"context"
	"net/url"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// ListTraces calls GET /traces.
func (c *Client) ListTraces(ctx context.Context, q url.Values) (*models.Paginated[models.TraceListItem], error) {
	return doDecode[models.Paginated[models.TraceListItem]](ctx, c, "GET", "/traces", q, nil)
}

// GetTrace calls GET /traces/{trace_id}, returning the trace with spans inlined.
func (c *Client) GetTrace(ctx context.Context, traceID string) (*models.TraceResponse, error) {
	return doDecode[models.TraceResponse](ctx, c, "GET", "/traces/"+url.PathEscape(traceID), nil, nil)
}
