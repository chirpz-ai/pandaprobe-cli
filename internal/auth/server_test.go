package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCallbackEscapesErrorDescription ensures a malicious error_description
// cannot inject markup into the failure page rendered on the loopback.
func TestCallbackEscapesErrorDescription(t *testing.T) {
	state := "good-state"
	cs, err := newCallbackServer(state)
	require.NoError(t, err)
	cs.start()
	defer cs.shutdown()

	evil := "<script>alert(1)</script>"
	u := fmt.Sprintf("http://127.0.0.1:%d/callback?state=%s&error=denied&error_description=%s",
		cs.Port(), url.QueryEscape(state), url.QueryEscape(evil))
	resp, err := http.Get(u) //nolint:gosec // loopback test URL
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.NotContains(t, string(body), "<script>", "raw markup must not appear in the page")
	assert.Contains(t, string(body), "&lt;script&gt;", "markup must be HTML-escaped")

	// The handler still reports the failure to the waiter.
	res := <-cs.result
	require.Error(t, res.err)
}
