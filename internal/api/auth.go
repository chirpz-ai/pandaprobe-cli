package api

import (
	"context"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// ExchangeCode swaps a one-time login code plus its PKCE verifier for a freshly
// minted API key via POST /cli/auth/exchange. No API key is required to call it
// (the code+verifier is the proof), but it reuses the standard request path so
// debug logging, timeouts, and APIError/exit-code mapping all apply.
func (c *Client) ExchangeCode(ctx context.Context, code, verifier string) (*models.CLICredentials, error) {
	body := models.ExchangeRequest{Code: code, CodeVerifier: verifier}
	return doDecode[models.CLICredentials](ctx, c, "POST", "/cli/auth/exchange", nil, body)
}
