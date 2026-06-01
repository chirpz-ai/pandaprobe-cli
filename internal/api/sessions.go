package api

import (
	"context"
	"net/url"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// ListSessions calls GET /sessions.
func (c *Client) ListSessions(ctx context.Context, q url.Values) (*models.Paginated[models.SessionSummary], error) {
	return doDecode[models.Paginated[models.SessionSummary]](ctx, c, "GET", "/sessions", q, nil)
}

// GetSession calls GET /sessions/{session_id}, returning the session with its
// traces (each with spans inlined).
func (c *Client) GetSession(ctx context.Context, sessionID string, q url.Values) (*models.SessionDetail, error) {
	return doDecode[models.SessionDetail](ctx, c, "GET", "/sessions/"+url.PathEscape(sessionID), q, nil)
}
