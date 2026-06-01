package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

type stubExchanger struct {
	creds       *models.CLICredentials
	err         error
	gotCode     string
	gotVerifier string
}

func (s *stubExchanger) ExchangeCode(_ context.Context, code, verifier string) (*models.CLICredentials, error) {
	s.gotCode, s.gotVerifier = code, verifier
	return s.creds, s.err
}

// hitCallback simulates the browser/web app redirecting to the loopback.
func hitCallback(t *testing.T, rawLoginURL, code, state string) {
	t.Helper()
	u, err := url.Parse(rawLoginURL)
	require.NoError(t, err)
	port := u.Query().Get("port")
	require.NotEmpty(t, port)
	cb := fmt.Sprintf("http://127.0.0.1:%s/callback?code=%s&state=%s", port, url.QueryEscape(code), url.QueryEscape(state))
	resp, err := http.Get(cb) //nolint:gosec // loopback test URL
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func TestLoginHappyPath(t *testing.T) {
	ex := &stubExchanger{creds: &models.CLICredentials{APIKey: "sk_pp_minted", ProjectName: "proj", ExpiresAt: "2026-09-01T00:00:00Z"}}

	var challenge string
	open := func(rawURL string) error {
		u, _ := url.Parse(rawURL)
		q := u.Query()
		challenge = q.Get("code_challenge")
		assert.Equal(t, "S256", q.Get("code_challenge_method"))
		assert.Equal(t, "/cli-login", u.Path)
		assert.NotEmpty(t, q.Get("state"))
		hitCallback(t, rawURL, "THE_CODE", q.Get("state"))
		return nil
	}

	creds, err := Login(context.Background(), ex, Options{
		AuthURL: "https://app.example.com",
		Label:   "test-host",
		Open:    open,
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, "sk_pp_minted", creds.APIKey)
	assert.Equal(t, "proj", creds.ProjectName)
	assert.Equal(t, "THE_CODE", ex.gotCode)
	// PKCE end to end: the verifier the exchanger received must hash to the
	// challenge the browser was sent.
	assert.Equal(t, challenge, Challenge(ex.gotVerifier))
}

func TestLoginRejectsBadState(t *testing.T) {
	ex := &stubExchanger{creds: &models.CLICredentials{APIKey: "x"}}
	open := func(rawURL string) error {
		hitCallback(t, rawURL, "CODE", "wrong-state")
		return nil
	}
	_, err := Login(context.Background(), ex, Options{AuthURL: "https://app.example.com", Open: open, Timeout: 5 * time.Second})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state mismatch")
	assert.Empty(t, ex.gotCode, "exchange must not run on state mismatch")
}

func TestLoginTimeout(t *testing.T) {
	ex := &stubExchanger{creds: &models.CLICredentials{APIKey: "x"}}
	open := func(string) error { return nil } // never completes the callback
	_, err := Login(context.Background(), ex, Options{AuthURL: "https://app.example.com", Open: open, Timeout: 150 * time.Millisecond})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestBuildLoginURLInvalid(t *testing.T) {
	_, err := buildLoginURL("not-a-url", 1234, "s", "c", "l", "v")
	assert.Error(t, err)
}
