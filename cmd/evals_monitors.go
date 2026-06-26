package cmd

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func newEvalsMonitorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitors",
		Short: "Create and manage evaluation monitors",
	}
	cmd.AddCommand(
		newEvalsMonitorsCreateCmd(),
		newEvalsMonitorsListCmd(),
		newEvalsMonitorsGetCmd(),
		newEvalsMonitorsUpdateCmd(),
		newEvalsMonitorsDeleteCmd(),
		newEvalsMonitorsPauseCmd(),
		newEvalsMonitorsResumeCmd(),
		newEvalsMonitorsRunsCmd(),
		newEvalsMonitorsTriggerCmd(),
	)
	return cmd
}

// validateCadence accepts a predefined cadence (every_6h/daily/weekly) or a
// "cron:" prefixed 5-field cron expression. The server performs the
// authoritative validation; this catches obvious client-side mistakes.
func validateCadence(c string) error {
	switch c {
	case "every_6h", "daily", "weekly":
		return nil
	}
	if expr, ok := strings.CutPrefix(c, "cron:"); ok {
		if len(strings.Fields(expr)) == 5 {
			return nil
		}
		return exitcode.New(exitcode.Validation,
			"invalid --cadence %q: cron expression must have 5 space-separated fields", c)
	}
	return exitcode.New(exitcode.Validation,
		"invalid --cadence %q: must be every_6h, daily, weekly, or cron:<5-field expression>", c)
}

func newEvalsMonitorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create an evaluation monitor from filters",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return exitcode.New(exitcode.Validation, "--name is required")
			}
			metrics, _ := cmd.Flags().GetStringSlice("metrics")
			if len(metrics) == 0 {
				return exitcode.New(exitcode.Validation, "at least one --metrics value is required")
			}
			cadence, _ := cmd.Flags().GetString("cadence")
			if cadence == "" {
				return exitcode.New(exitcode.Validation, "--cadence is required")
			}
			if err := validateCadence(cadence); err != nil {
				return err
			}

			body := models.CreateMonitorRequest{
				Name:          name,
				TargetType:    strings.ToUpper(target),
				Metrics:       metrics,
				Cadence:       cadence,
				SamplingRate:  floatPtrFlag(cmd, "sampling-rate"),
				Model:         strPtrFlag(cmd, "model"),
				OnlyIfChanged: boolPtrFlag(cmd, "only-if-changed"),
			}
			if target == targetSession {
				weights, err := signalWeights(cmd)
				if err != nil {
					return err
				}
				body.SignalWeights = weights
				body.Filters = &models.SessionEvalRunFilters{
					DateFrom:      strPtrFlag(cmd, "date-from"),
					DateTo:        strPtrFlag(cmd, "date-to"),
					UserID:        strPtrFlag(cmd, "user-id"),
					HasError:      boolPtrFlag(cmd, "has-error"),
					Tags:          sliceFlag(cmd, "tags"),
					MinTraceCount: intPtrFlag(cmd, "min-trace-count"),
				}
			} else {
				status, _ := cmd.Flags().GetString("status")
				if err := validateEnum("status", status, models.TraceStatusValues()); err != nil {
					return err
				}
				body.Filters = &models.EvalRunFilters{
					DateFrom:  strPtrFlag(cmd, "date-from"),
					DateTo:    strPtrFlag(cmd, "date-to"),
					Status:    strPtrFlag(cmd, "status"),
					SessionID: strPtrFlag(cmd, "session-id"),
					UserID:    strPtrFlag(cmd, "user-id"),
					Tags:      sliceFlag(cmd, "tags"),
					Name:      strPtrFlag(cmd, "filter-name"),
				}
			}

			res, err := app.client.CreateMonitor(cmd.Context(), body)
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
	cmd.Flags().String("name", "", "Human-readable monitor name (required)")
	cmd.Flags().StringSlice("metrics", nil, "Metric names to run (required, comma-separated)")
	cmd.Flags().String("cadence", "", "Schedule: every_6h, daily, weekly, or cron:<5-field expr> (required)")
	cmd.Flags().Float64("sampling-rate", 0, "Fraction of targets to evaluate (0-1)")
	cmd.Flags().String("model", "", "Override the LLM judge model")
	cmd.Flags().Bool("only-if-changed", true, "Only run when matching targets changed since the last run")
	// Trace filters.
	cmd.Flags().String("date-from", "", "Filter: targets on/after this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("date-to", "", "Filter: targets on/before this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("status", "", "Filter (trace): trace status")
	cmd.Flags().String("session-id", "", "Filter (trace): session ID")
	cmd.Flags().String("user-id", "", "Filter: user ID")
	cmd.Flags().StringSlice("tags", nil, "Filter: tags (any match)")
	cmd.Flags().String("filter-name", "", "Filter (trace): substring match on trace name")
	// Session filters.
	cmd.Flags().Bool("has-error", false, "Filter (session): only sessions with/without errors")
	cmd.Flags().Int("min-trace-count", 0, "Filter (session): minimum traces per session")
	cmd.Flags().String("signal-weights", "", "Session only: JSON object overriding signal weights")
	return cmd
}

func newEvalsMonitorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List evaluation monitors",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			q, err := newQuery(cmd).
				pagination(evalMaxLimit).
				enum("status", "status", models.MonitorStatusValues()).
				build()
			if err != nil {
				return err
			}
			res, err := app.client.ListMonitors(cmd.Context(), q)
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().String("status", "", "Filter by monitor status (ACTIVE, PAUSED)")
	return cmd
}

func newEvalsMonitorsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "get <monitor_id>",
		Short:       "Get an evaluation monitor",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			res, err := app.client.GetMonitor(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
}

func newEvalsMonitorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <monitor_id>",
		Short:       "Update an evaluation monitor (partial)",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			if f := cmd.Flags().Lookup("cadence"); f != nil && f.Changed {
				cadence, _ := cmd.Flags().GetString("cadence")
				if err := validateCadence(cadence); err != nil {
					return err
				}
			}
			body := models.UpdateMonitorRequest{
				Name:          strPtrFlag(cmd, "name"),
				Metrics:       sliceFlag(cmd, "metrics"),
				SamplingRate:  floatPtrFlag(cmd, "sampling-rate"),
				Model:         strPtrFlag(cmd, "model"),
				Cadence:       strPtrFlag(cmd, "cadence"),
				OnlyIfChanged: boolPtrFlag(cmd, "only-if-changed"),
			}
			if f := cmd.Flags().Lookup("filters"); f != nil && f.Changed {
				raw, _ := cmd.Flags().GetString("filters")
				var obj map[string]any
				if err := json.Unmarshal([]byte(raw), &obj); err != nil {
					return exitcode.New(exitcode.Validation, "invalid --filters: must be a JSON object")
				}
				body.Filters = json.RawMessage(raw)
			}
			weights, err := signalWeights(cmd)
			if err != nil {
				return err
			}
			body.SignalWeights = weights

			res, err := app.client.UpdateMonitor(cmd.Context(), args[0], body)
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
	cmd.Flags().String("name", "", "New monitor name")
	cmd.Flags().StringSlice("metrics", nil, "New metric names (comma-separated)")
	cmd.Flags().String("cadence", "", "New schedule: every_6h, daily, weekly, or cron:<5-field expr>")
	cmd.Flags().Float64("sampling-rate", 0, "New sampling rate (0-1)")
	cmd.Flags().String("model", "", "New LLM judge model override")
	cmd.Flags().Bool("only-if-changed", true, "Only run when matching targets changed since the last run")
	cmd.Flags().String("filters", "", "Replacement filters as a JSON object")
	cmd.Flags().String("signal-weights", "", "Session only: JSON object overriding signal weights")
	return cmd
}

func newEvalsMonitorsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "delete <monitor_id>",
		Short:       "Delete an evaluation monitor",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			if err := app.client.DeleteMonitor(cmd.Context(), args[0]); err != nil {
				return err
			}
			return app.writer.Render(map[string]string{"status": "deleted", "id": args[0]})
		},
	}
}

func newEvalsMonitorsPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "pause <monitor_id>",
		Short:       "Pause an evaluation monitor",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			res, err := app.client.PauseMonitor(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
}

func newEvalsMonitorsResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "resume <monitor_id>",
		Short:       "Resume a paused evaluation monitor",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			res, err := app.client.ResumeMonitor(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
}

func newEvalsMonitorsRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "runs <monitor_id>",
		Short:       "List the evaluation runs produced by a monitor",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			q, err := newQuery(cmd).pagination(evalMaxLimit).build()
			if err != nil {
				return err
			}
			res, err := app.client.MonitorRuns(cmd.Context(), args[0], q)
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	return cmd
}

func newEvalsMonitorsTriggerCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "trigger <monitor_id>",
		Short:       "Trigger an immediate run of a monitor (server rate-limited)",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			res, err := app.client.TriggerMonitor(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
}
