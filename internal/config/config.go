// Package config resolves runtime configuration from (in precedence order)
// command-line flags, PANDAPROBE_* environment variables, the config file at
// ~/.pandaprobe/config.yaml, and built-in defaults.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

// Config keys. These are the canonical (snake_case) names used in the config
// file, as Viper keys, and (uppercased, prefixed) as environment variables.
const (
	KeyAPIKey      = "api_key"
	KeyProjectName = "project_name"
	KeyEndpoint    = "endpoint"
	KeyFormat      = "format"
	KeyTimeout     = "timeout"
)

// EnvPrefix is prepended to config keys to form environment variable names,
// e.g. api_key -> PANDAPROBE_API_KEY.
const EnvPrefix = "PANDAPROBE"

// Defaults.
const (
	DefaultEndpoint = "https://api.pandaprobe.com"
	DefaultFormat   = "json"
	DefaultTimeout  = 30 * time.Second
)

// SettableKeys are the keys accepted by `config set` / `config get`.
var SettableKeys = []string{KeyAPIKey, KeyProjectName, KeyEndpoint, KeyFormat}

// Config is the fully resolved runtime configuration.
type Config struct {
	APIKey      string
	ProjectName string
	Endpoint    string
	Format      string
	Timeout     time.Duration

	// Global behavioral flags (not persisted to the config file).
	NoColor bool
	Verbose bool
	Debug   bool
}

// Source identifies where a resolved value originated.
type Source string

const (
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceFile    Source = "file"
	SourceDefault Source = "default"
)

// DefaultDir returns the directory holding the config file (~/.pandaprobe).
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".pandaprobe"), nil
}

// DefaultPath returns the default config file path (~/.pandaprobe/config.yaml).
func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// NewViper builds a Viper instance wired for the CLI's precedence chain. The
// caller is responsible for binding flags (via BindPFlag) before reading
// values. If cfgFile is empty the default path is used; a missing default file
// is not an error.
func NewViper(cfgFile string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetDefault(KeyEndpoint, DefaultEndpoint)
	v.SetDefault(KeyFormat, DefaultFormat)
	v.SetDefault(KeyTimeout, int(DefaultTimeout.Seconds()))

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, exitcode.New(exitcode.Validation, "read config file %q: %v", cfgFile, err)
		}
		return v, nil
	}

	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(*os.PathError); !ok && !os.IsNotExist(err) {
			// A genuinely malformed (but present) file should surface.
			if _, statErr := os.Stat(path); statErr == nil {
				return nil, exitcode.New(exitcode.Validation, "read config file %q: %v", path, err)
			}
		}
		// Missing default file is fine.
	}
	return v, nil
}

// Load reads the resolved configuration from a prepared Viper instance.
func Load(v *viper.Viper) (*Config, error) {
	c := &Config{
		APIKey:      v.GetString(KeyAPIKey),
		ProjectName: v.GetString(KeyProjectName),
		Endpoint:    v.GetString(KeyEndpoint),
		Format:      v.GetString(KeyFormat),
	}

	secs := v.GetInt(KeyTimeout)
	if secs <= 0 {
		c.Timeout = DefaultTimeout
	} else {
		c.Timeout = time.Duration(secs) * time.Second
	}
	return c, nil
}

// Validate performs pre-flight checks. When requiresAuth is true the API key
// and project name must be present (commands that hit the API). Endpoint and
// format are always validated.
func (c *Config) Validate(requiresAuth bool) error {
	if c.Format != "json" && c.Format != "table" {
		return exitcode.New(exitcode.Validation,
			"invalid format %q: must be \"json\" or \"table\"", c.Format)
	}
	if strings.TrimSpace(c.Endpoint) == "" {
		return exitcode.New(exitcode.Validation,
			"no endpoint configured: set --endpoint, PANDAPROBE_ENDPOINT, or run `pandaprobe config set endpoint <url>`")
	}
	if u, err := url.Parse(c.Endpoint); err != nil || u.Scheme == "" || u.Host == "" {
		return exitcode.New(exitcode.Validation, "invalid endpoint %q: must be an absolute URL", c.Endpoint)
	}
	if requiresAuth {
		if strings.TrimSpace(c.APIKey) == "" {
			return exitcode.New(exitcode.Auth,
				"no API key configured: set --api-key, PANDAPROBE_API_KEY, or run `pandaprobe config set api_key <key>`")
		}
		if strings.TrimSpace(c.ProjectName) == "" {
			return exitcode.New(exitcode.Validation,
				"no project configured: set --project, PANDAPROBE_PROJECT_NAME, or run `pandaprobe config set project_name <name>`")
		}
	}
	return nil
}

// ResolveSource reports where the effective value of key came from, mirroring
// the precedence chain: an explicitly-changed flag wins, then environment, then
// the config file, then defaults. The caller supplies flagChanged because
// Viper's IsSet cannot distinguish a flag-set value from a SetDefault value.
func ResolveSource(v *viper.Viper, key string, flagChanged bool) Source {
	if flagChanged {
		return SourceFlag
	}
	envName := EnvPrefix + "_" + strings.ToUpper(key)
	if _, ok := os.LookupEnv(envName); ok {
		return SourceEnv
	}
	if v.InConfig(key) {
		return SourceFile
	}
	return SourceDefault
}

// SetValue persists a single key/value to the config file at path, creating the
// directory (0700) and file (0600) if needed. It validates the key and value.
func SetValue(path, key, value string) error {
	allowed := false
	for _, k := range SettableKeys {
		if k == key {
			allowed = true
			break
		}
	}
	if !allowed {
		return exitcode.New(exitcode.Validation,
			"unknown config key %q: valid keys are %s", key, strings.Join(SettableKeys, ", "))
	}
	if key == KeyFormat && value != "json" && value != "table" {
		return exitcode.New(exitcode.Validation, "invalid format %q: must be \"json\" or \"table\"", value)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return exitcode.New(exitcode.General, "create config dir %q: %v", dir, err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	if _, err := os.Stat(path); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return exitcode.New(exitcode.Validation, "read config file %q: %v", path, err)
		}
	}
	v.Set(key, value)
	if err := v.WriteConfigAs(path); err != nil {
		return exitcode.New(exitcode.General, "write config file %q: %v", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return exitcode.New(exitcode.General, "secure config file %q: %v", path, err)
	}
	return nil
}
