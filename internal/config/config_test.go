package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBoundViper builds a Viper with a flag set bound, mirroring how root.go
// wires things, so precedence tests are faithful.
func newBoundViper(t *testing.T, cfgFile string) (*viper.Viper, *pflag.FlagSet) {
	t.Helper()
	v, err := NewViper(cfgFile)
	require.NoError(t, err)

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("api-key", "", "")
	fs.String("project", "", "")
	fs.String("endpoint", "", "")
	fs.String("format", "", "")

	require.NoError(t, v.BindPFlag(KeyAPIKey, fs.Lookup("api-key")))
	require.NoError(t, v.BindPFlag(KeyProjectName, fs.Lookup("project")))
	require.NoError(t, v.BindPFlag(KeyEndpoint, fs.Lookup("endpoint")))
	require.NoError(t, v.BindPFlag(KeyFormat, fs.Lookup("format")))
	return v, fs
}

func TestDefaults(t *testing.T) {
	v, _ := newBoundViper(t, "")
	cfg, err := Load(v)
	require.NoError(t, err)
	assert.Equal(t, DefaultEndpoint, cfg.Endpoint)
	assert.Equal(t, DefaultFormat, cfg.Format)
	assert.Equal(t, DefaultTimeout, cfg.Timeout)
	assert.Equal(t, SourceDefault, ResolveSource(v, KeyEndpoint, false))
}

func TestEnvOverridesDefault(t *testing.T) {
	t.Setenv("PANDAPROBE_ENDPOINT", "https://env.example.com")
	t.Setenv("PANDAPROBE_API_KEY", "sk_pp_env")
	v, _ := newBoundViper(t, "")
	cfg, err := Load(v)
	require.NoError(t, err)
	assert.Equal(t, "https://env.example.com", cfg.Endpoint)
	assert.Equal(t, "sk_pp_env", cfg.APIKey)
	assert.Equal(t, SourceEnv, ResolveSource(v, KeyEndpoint, false))
}

func TestFileOverridesDefaultAndFlagOverridesAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("endpoint: https://file.example.com\napi_key: sk_pp_file\nproject_name: from-file\n"), 0o600))

	v, fs := newBoundViper(t, path)
	cfg, err := Load(v)
	require.NoError(t, err)
	assert.Equal(t, "https://file.example.com", cfg.Endpoint)
	assert.Equal(t, "from-file", cfg.ProjectName)
	assert.Equal(t, SourceFile, ResolveSource(v, KeyEndpoint, false))

	// A flag must win over the file.
	require.NoError(t, fs.Set("endpoint", "https://flag.example.com"))
	cfg, err = Load(v)
	require.NoError(t, err)
	assert.Equal(t, "https://flag.example.com", cfg.Endpoint)
	assert.Equal(t, SourceFlag, ResolveSource(v, KeyEndpoint, fs.Lookup("endpoint").Changed))
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("endpoint: https://file.example.com\n"), 0o600))
	t.Setenv("PANDAPROBE_ENDPOINT", "https://env.example.com")

	v, _ := newBoundViper(t, path)
	cfg, err := Load(v)
	require.NoError(t, err)
	assert.Equal(t, "https://env.example.com", cfg.Endpoint)
	assert.Equal(t, SourceEnv, ResolveSource(v, KeyEndpoint, false))
}

func TestTimeoutFromEnv(t *testing.T) {
	t.Setenv("PANDAPROBE_TIMEOUT", "60")
	v, _ := newBoundViper(t, "")
	cfg, err := Load(v)
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		requiresAuth bool
		wantErr      bool
	}{
		{"ok no auth", Config{Endpoint: DefaultEndpoint, Format: "json"}, false, false},
		{"bad format", Config{Endpoint: DefaultEndpoint, Format: "xml"}, false, true},
		{"empty endpoint", Config{Endpoint: "", Format: "json"}, false, true},
		{"bad endpoint", Config{Endpoint: "not a url", Format: "json"}, false, true},
		{"missing key", Config{Endpoint: DefaultEndpoint, Format: "json"}, true, true},
		{"ok auth", Config{Endpoint: DefaultEndpoint, Format: "json", APIKey: "k", ProjectName: "p"}, true, false},
		{"missing project", Config{Endpoint: DefaultEndpoint, Format: "json", APIKey: "k"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate(tt.requiresAuth)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetValueRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	require.NoError(t, SetValue(path, KeyAPIKey, "sk_pp_secret"))
	require.NoError(t, SetValue(path, KeyProjectName, "proj"))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	v := viper.New()
	v.SetConfigFile(path)
	require.NoError(t, v.ReadInConfig())
	assert.Equal(t, "sk_pp_secret", v.GetString(KeyAPIKey))
	assert.Equal(t, "proj", v.GetString(KeyProjectName))

	assert.Error(t, SetValue(path, "bogus", "x"))
	assert.Error(t, SetValue(path, KeyFormat, "xml"))
}
