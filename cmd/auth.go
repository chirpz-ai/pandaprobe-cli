package cmd

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/auth"
	"github.com/chirpz-ai/pandaprobe-cli/internal/config"
	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/output"
	"github.com/chirpz-ai/pandaprobe-cli/internal/version"
)

// openBrowser is a test seam; tests replace it to simulate the web app.
var openBrowser = auth.OpenBrowser

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with PandaProbe Cloud",
		Long:  "Log in via the browser to mint and store an API key automatically.\nFor local/self-hosted PandaProbe with auth disabled, use `config set` instead.",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd(), newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var noBrowser bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in through the browser and store a minted API key",
		Args:  cobra.NoArgs,
		// Not annotated requiresAuth: this command acquires the credentials.
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)

			authURL := app.cfg.AuthURL
			if err := validateURL("auth URL", authURL); err != nil {
				return err
			}
			if err := validateURL("endpoint", app.cfg.Endpoint); err != nil {
				return err
			}

			creds, err := auth.Login(cmd.Context(), app.client, auth.Options{
				AuthURL:    authURL,
				Label:      hostnameLabel(),
				CLIVersion: version.Version,
				NoBrowser:  noBrowser,
				Timeout:    timeout,
				Open:       openBrowser,
				Progress:   cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}

			// Persist the minted credentials to the config file.
			path, perr := configPath(cmd)
			if perr != nil {
				return perr
			}
			if err := config.SetValue(path, config.KeyAPIKey, creds.APIKey); err != nil {
				return err
			}
			if creds.ProjectName != "" {
				if err := config.SetValue(path, config.KeyProjectName, creds.ProjectName); err != nil {
					return err
				}
			}
			if creds.Endpoint != "" && creds.Endpoint != app.cfg.Endpoint {
				if err := config.SetValue(path, config.KeyEndpoint, creds.Endpoint); err != nil {
					return err
				}
			}

			endpoint := creds.Endpoint
			if endpoint == "" {
				endpoint = app.cfg.Endpoint
			}
			return app.writer.Render(map[string]any{
				"logged_in":   true,
				"project":     creds.ProjectName,
				"api_key":     output.MaskSecret(creds.APIKey),
				"endpoint":    endpoint,
				"expires_at":  creds.ExpiresAt,
				"config_file": path,
			})
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the login URL instead of opening a browser")
	cmd.Flags().DurationVar(&timeout, "timeout", auth.DefaultTimeout, "How long to wait for browser authorization")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether the CLI is logged in",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			return app.writer.Render(map[string]any{
				"logged_in":    strings.TrimSpace(app.cfg.APIKey) != "",
				"api_key":      output.MaskSecret(app.cfg.APIKey),
				"project_name": app.cfg.ProjectName,
				"endpoint":     app.cfg.Endpoint,
				"auth_url":     app.cfg.AuthURL,
			})
		},
	}
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials from the config file",
		Long:  "Removes the stored API key and project from the local config file.\nTo revoke the key server-side, delete it from the PandaProbe dashboard.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			if err := config.UnsetValue(path, config.KeyAPIKey, config.KeyProjectName); err != nil {
				return err
			}
			return app.writer.Render(map[string]any{"logged_out": true, "config_file": path})
		},
	}
	return cmd
}

func validateURL(name, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return exitcode.New(exitcode.Validation, "no %s configured", name)
	}
	if u, err := url.Parse(raw); err != nil || u.Scheme == "" || u.Host == "" {
		return exitcode.New(exitcode.Validation, "invalid %s %q: must be an absolute URL", name, raw)
	}
	return nil
}

func hostnameLabel() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "pandaprobe-cli"
	}
	return h
}
