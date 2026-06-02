// Package auth implements the `pandaprobe auth login` browser flow: an OAuth
// 2.0 Authorization-Code + PKCE handshake that mints an API key server-side.
// The raw key never transits the browser or the loopback redirect — only a
// single-use code does, which the CLI exchanges directly over HTTPS.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// randomURLSafe returns n random bytes encoded as an unpadded base64url string.
func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomState returns an opaque CSRF nonce echoed through the browser and
// verified on the loopback callback.
func RandomState() (string, error) { return randomURLSafe(32) }

// RandomVerifier returns a PKCE code verifier (high-entropy, URL-safe).
func RandomVerifier() (string, error) { return randomURLSafe(32) }

// Challenge derives the PKCE S256 code challenge from a verifier:
// BASE64URL(SHA256(verifier)).
func Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
