package models

// ExchangeRequest is the body of POST /cli/auth/exchange. The one-time code plus
// the PKCE verifier are the only proof of identity (no auth header is sent).
type ExchangeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

// CLICredentials is the result of a successful code exchange: a freshly minted
// API key plus the context the CLI persists to its config.
type CLICredentials struct {
	APIKey      string `json:"api_key"`
	ProjectName string `json:"project_name"`
	Endpoint    string `json:"endpoint"`
	OrgID       string `json:"org_id"`
	KeyID       string `json:"key_id"`
	KeyPrefix   string `json:"key_prefix"`
	ExpiresAt   string `json:"expires_at"`
}
