package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeIsS256OfVerifier(t *testing.T) {
	v, err := RandomVerifier()
	require.NoError(t, err)

	got := Challenge(v)
	sum := sha256.Sum256([]byte(v))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	assert.Equal(t, want, got)
	// base64url must not contain padding or unsafe chars.
	assert.NotContains(t, got, "=")
	assert.NotContains(t, got, "+")
	assert.NotContains(t, got, "/")
}

func TestRandomValuesAreUniqueAndNonEmpty(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		s, err := RandomState()
		require.NoError(t, err)
		assert.NotEmpty(t, s)
		assert.False(t, seen[s], "state collision")
		seen[s] = true

		v, err := RandomVerifier()
		require.NoError(t, err)
		assert.NotEmpty(t, v)
	}
}
