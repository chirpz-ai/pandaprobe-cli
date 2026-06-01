// Package cmd wires the Cobra command tree. Command files are intentionally
// thin: they parse and validate flags, call internal/api, and hand results to
// internal/output. No HTTP or JSON logic lives here.
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chirpz-ai/pandaprobe-cli/internal/api"
	"github.com/chirpz-ai/pandaprobe-cli/internal/config"
	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/output"
)

// annotationRequiresAuth marks commands that call the authenticated API and
// therefore need a valid API key, project, and endpoint before running.
const annotationRequiresAuth = "requiresAuth"

// appContext holds the per-invocation dependencies built in PersistentPreRunE.
type appContext struct {
	cfg    *config.Config
	writer *output.Writer
	client *api.Client
	v      *viper.Viper
	cmd    *cobra.Command
}

type ctxKey struct{}

// appFrom retrieves the appContext stored on the command's context.
func appFrom(cmd *cobra.Command) *appContext {
	if a, ok := cmd.Context().Value(ctxKey{}).(*appContext); ok {
		return a
	}
	return &appContext{}
}

func newRootCmd(app *appContext) *cobra.Command {
	root := &cobra.Command{
		Use:           "pandaprobe",
		Short:         "Agent-first CLI for the PandaProbe LLM observability platform",
		Long:          "pandaprobe inspects traces, sessions, spans, scores and evaluation runs from the\nPandaProbe backend. JSON output is the default; pass --format table for humans.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := root.PersistentFlags()
	pf.String("api-key", "", "PandaProbe API key (env PANDAPROBE_API_KEY)")
	pf.String("project", "", "PandaProbe project name (env PANDAPROBE_PROJECT_NAME)")
	pf.String("endpoint", "", "API endpoint URL (env PANDAPROBE_ENDPOINT)")
	pf.String("format", "", "Output format: json or table (default json)")
	pf.String("config", "", "Path to config file (default ~/.pandaprobe/config.yaml)")
	pf.Bool("verbose", false, "Log request/response summaries to stderr")
	pf.Bool("debug", false, "Log full HTTP request/response details to stderr")
	pf.Bool("no-color", false, "Disable color in table output")

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return setupApp(cmd, app)
	}

	root.AddCommand(
		newVersionCmd(),
		newCompletionCmd(),
		newConfigCmd(),
		newTracesCmd(),
		newSessionsCmd(),
		newEvalsCmd(),
	)
	return root
}

// setupApp builds config, output writer, and API client, then runs pre-flight
// validation. It populates app so the caller can render errors uniformly.
func setupApp(cmd *cobra.Command, app *appContext) error {
	flags := cmd.Flags()
	cfgFile, _ := flags.GetString("config")

	v, err := config.NewViper(cfgFile)
	if err != nil {
		return err
	}
	bind := func(key, flag string) error { return v.BindPFlag(key, flags.Lookup(flag)) }
	if err := bind(config.KeyAPIKey, "api-key"); err != nil {
		return err
	}
	if err := bind(config.KeyProjectName, "project"); err != nil {
		return err
	}
	if err := bind(config.KeyEndpoint, "endpoint"); err != nil {
		return err
	}
	if err := bind(config.KeyFormat, "format"); err != nil {
		return err
	}

	cfg, err := config.Load(v)
	if err != nil {
		return err
	}
	cfg.NoColor, _ = flags.GetBool("no-color")
	cfg.Verbose, _ = flags.GetBool("verbose")
	cfg.Debug, _ = flags.GetBool("debug")

	app.cfg = cfg
	app.v = v
	app.cmd = cmd
	app.writer = output.New(cmd.OutOrStdout(), cmd.ErrOrStderr(), cfg.Format, cfg.NoColor)

	requiresAuth := cmd.Annotations[annotationRequiresAuth] == "true"
	if err := cfg.Validate(requiresAuth); err != nil {
		return err
	}

	var debugOut *os.File
	if cfg.Debug || cfg.Verbose {
		debugOut = os.Stderr
	}
	app.client = api.New(cfg, debugOut)

	ctx := context.WithValue(cmd.Context(), ctxKey{}, app)
	cmd.SetContext(ctx)
	return nil
}

// Execute builds and runs the root command, rendering any error in the active
// format and returning the resolved process exit code.
func Execute() exitcode.Code {
	app := &appContext{}
	root := newRootCmd(app)
	return executeRoot(root, app)
}

// executeRoot runs a prepared root command and renders any error uniformly. It
// is separated from Execute so tests can supply their own root with captured
// output streams.
func executeRoot(root *cobra.Command, app *appContext) exitcode.Code {
	err := root.Execute()
	if err == nil {
		return exitcode.OK
	}
	w := app.writer
	if w == nil {
		// Error occurred before the writer was built (e.g. flag parse error).
		w = output.New(root.OutOrStdout(), root.ErrOrStderr(), string(output.FormatJSON), false)
	}
	_ = w.RenderError(err)
	return exitcode.From(err)
}
