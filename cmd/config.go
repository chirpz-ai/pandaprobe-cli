package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/config"
	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/output"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration (~/.pandaprobe/config.yaml)",
	}
	cmd.AddCommand(newConfigSetCmd(), newConfigGetCmd(), newConfigShowCmd(), newConfigPathCmd())
	return cmd
}

// configKeyFlag maps a config key to the persistent flag that can set it.
var configKeyFlag = map[string]string{
	config.KeyAPIKey:      "api-key",
	config.KeyProjectName: "project",
	config.KeyEndpoint:    "endpoint",
	config.KeyAuthURL:     "auth-url",
	config.KeyFormat:      "format",
}

func configPath(cmd *cobra.Command) (string, error) {
	if p, _ := cmd.Flags().GetString("config"); p != "" {
		return p, nil
	}
	return config.DefaultPath()
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value (keys: api_key, project_name, endpoint, format)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			key, value := args[0], args[1]
			if err := config.SetValue(path, key, value); err != nil {
				return err
			}
			shown := value
			if key == config.KeyAPIKey {
				shown = output.MaskSecret(value)
			}
			return appFrom(cmd).writer.Render(map[string]string{
				"key": key, "value": shown, "config_file": path,
			})
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	var reveal bool
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a single resolved config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			key := args[0]
			val, ok := resolvedValue(app, key)
			if !ok {
				return exitcode.New(exitcode.Validation, "unknown config key %q", key)
			}
			if key == config.KeyAPIKey && !reveal {
				val = output.MaskSecret(val)
			}
			return app.writer.Render(map[string]string{"key": key, "value": val})
		},
	}
	cmd.Flags().BoolVar(&reveal, "reveal-secrets", false, "Show the full API key value")
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var reveal bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the effective configuration and where each value came from",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			path, _ := configPath(cmd)

			keys := []string{config.KeyAPIKey, config.KeyProjectName, config.KeyEndpoint, config.KeyAuthURL, config.KeyFormat, config.KeyTimeout}
			if app.writer.Format() == output.FormatTable {
				m := map[string]string{"config_file": path}
				for _, k := range keys {
					v, _ := resolvedValue(app, k)
					if k == config.KeyAPIKey && !reveal {
						v = output.MaskSecret(v)
					}
					m[k] = v + " (" + string(sourceOf(app, k)) + ")"
				}
				return app.writer.Render(m)
			}

			type entry struct {
				Value  string `json:"value"`
				Source string `json:"source"`
			}
			out := map[string]any{"config_file": path}
			for _, k := range keys {
				v, _ := resolvedValue(app, k)
				if k == config.KeyAPIKey && !reveal {
					v = output.MaskSecret(v)
				}
				out[k] = entry{Value: v, Source: string(sourceOf(app, k))}
			}
			return app.writer.Render(out)
		},
	}
	cmd.Flags().BoolVar(&reveal, "reveal-secrets", false, "Show the full API key value")
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path and whether it exists",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			exists := false
			if _, statErr := os.Stat(path); statErr == nil {
				exists = true
			}
			return appFrom(cmd).writer.Render(map[string]any{"path": path, "exists": exists})
		},
	}
}

func resolvedValue(app *appContext, key string) (string, bool) {
	switch key {
	case config.KeyAPIKey:
		return app.cfg.APIKey, true
	case config.KeyProjectName:
		return app.cfg.ProjectName, true
	case config.KeyEndpoint:
		return app.cfg.Endpoint, true
	case config.KeyAuthURL:
		return app.cfg.AuthURL, true
	case config.KeyFormat:
		return app.cfg.Format, true
	case config.KeyTimeout:
		return app.cfg.Timeout.String(), true
	default:
		return "", false
	}
}

func sourceOf(app *appContext, key string) config.Source {
	changed := false
	if flag, ok := configKeyFlag[key]; ok {
		if f := app.cmd.Flags().Lookup(flag); f != nil {
			changed = f.Changed
		}
	}
	return config.ResolveSource(app.v, key, changed)
}
