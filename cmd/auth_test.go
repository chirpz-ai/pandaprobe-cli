package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

func TestAuthLoginEndToEnd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Mock backend: the exchange endpoint mints a key.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/cli/auth/exchange", r.URL.Path)
		require.Equal(t, "POST", r.Method)
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.NotEmpty(t, body["code"])
		require.NotEmpty(t, body["code_verifier"])
		_, _ = w.Write([]byte(`{"api_key":"sk_pp_fromlogin","project_name":"login-proj","endpoint":"","org_id":"o1","key_id":"k1","key_prefix":"sk_pp_","expires_at":"2026-09-01T00:00:00Z"}`))
	}))
	defer srv.Close()
	t.Setenv("PANDAPROBE_ENDPOINT", srv.URL)

	// Simulate the browser: parse the login URL and hit the loopback callback.
	orig := openBrowser
	openBrowser = func(rawURL string) error {
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		q := u.Query()
		cb := fmt.Sprintf("http://127.0.0.1:%s/callback?code=ABC&state=%s", q.Get("port"), url.QueryEscape(q.Get("state")))
		resp, err := http.Get(cb) //nolint:gosec // loopback test URL
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		return nil
	}
	t.Cleanup(func() { openBrowser = orig })

	out, errOut, code := runCLI(t, "auth", "login")
	require.Equal(t, exitcode.OK, code, "stderr=%s", errOut)
	assert.Contains(t, out, `"logged_in": true`)
	assert.Contains(t, out, `"project": "login-proj"`)
	assert.Contains(t, out, "sk_pp_****ogin") // masked, not raw
	assert.NotContains(t, out, "sk_pp_fromlogin")

	// The credentials must be persisted to the config file.
	v := viper.New()
	v.SetConfigFile(filepath.Join(home, ".pandaprobe", "config.yaml"))
	require.NoError(t, v.ReadInConfig())
	assert.Equal(t, "sk_pp_fromlogin", v.GetString("api_key"))
	assert.Equal(t, "login-proj", v.GetString("project_name"))
}

func TestAuthStatusLoggedOutThenIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PANDAPROBE_API_KEY", "")
	out, _, code := runCLI(t, "auth", "status")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, `"logged_in": false`)

	t.Setenv("PANDAPROBE_API_KEY", "sk_pp_secretvalue99")
	out, _, code = runCLI(t, "auth", "status")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, `"logged_in": true`)
	assert.Contains(t, out, "sk_pp_****ue99")
	assert.NotContains(t, out, "secretvalue99")
}

func TestAuthLogoutClearsConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Seed a config via `config set`.
	_, _, code := runCLI(t, "config", "set", "api_key", "sk_pp_tobecleared")
	require.Equal(t, exitcode.OK, code)
	_, _, code = runCLI(t, "config", "set", "project_name", "p")
	require.Equal(t, exitcode.OK, code)

	out, _, code := runCLI(t, "auth", "logout")
	require.Equal(t, exitcode.OK, code)
	assert.Contains(t, out, `"logged_out": true`)

	data, err := os.ReadFile(filepath.Join(home, ".pandaprobe", "config.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "sk_pp_tobecleared")
	assert.NotContains(t, string(data), "project_name")
}
