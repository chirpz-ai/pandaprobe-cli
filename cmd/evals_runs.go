package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func newEvalsRunsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Create and inspect evaluation runs",
	}
	cmd.AddCommand(newEvalsRunsCreateCmd(), newEvalsRunsBatchCmd(), newEvalsRunsListCmd(),
		newEvalsRunsGetCmd(), newEvalsRunsScoresCmd())
	return cmd
}

func floatPtrFlag(cmd *cobra.Command, flag string) *float64 {
	f := cmd.Flags().Lookup(flag)
	if f == nil || !f.Changed {
		return nil
	}
	val, _ := cmd.Flags().GetFloat64(flag)
	return &val
}

func intPtrFlag(cmd *cobra.Command, flag string) *int {
	f := cmd.Flags().Lookup(flag)
	if f == nil || !f.Changed {
		return nil
	}
	val, _ := cmd.Flags().GetInt(flag)
	return &val
}

func boolPtrFlag(cmd *cobra.Command, flag string) *bool {
	f := cmd.Flags().Lookup(flag)
	if f == nil || !f.Changed {
		return nil
	}
	val, _ := cmd.Flags().GetBool(flag)
	return &val
}

// signalWeights parses the --signal-weights JSON flag into a map.
func signalWeights(cmd *cobra.Command) (map[string]float64, error) {
	raw, _ := cmd.Flags().GetString("signal-weights")
	if raw == "" {
		return nil, nil
	}
	var m map[string]float64
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, exitcode.New(exitcode.Validation, "invalid --signal-weights: must be a JSON object of name->number: %v", err)
	}
	return m, nil
}

func newEvalsRunsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create an evaluation run from filters",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			metrics, _ := cmd.Flags().GetStringSlice("metrics")
			if len(metrics) == 0 {
				return exitcode.New(exitcode.Validation, "at least one --metrics value is required")
			}
			if target == targetSession {
				weights, err := signalWeights(cmd)
				if err != nil {
					return err
				}
				body := models.CreateSessionEvalRunRequest{
					Name:          strPtrFlag(cmd, "name"),
					Metrics:       metrics,
					SamplingRate:  floatPtrFlag(cmd, "sampling-rate"),
					Model:         strPtrFlag(cmd, "model"),
					SignalWeights: weights,
					Filters: &models.SessionEvalRunFilters{
						DateFrom:      strPtrFlag(cmd, "date-from"),
						DateTo:        strPtrFlag(cmd, "date-to"),
						UserID:        strPtrFlag(cmd, "user-id"),
						HasError:      boolPtrFlag(cmd, "has-error"),
						Tags:          sliceFlag(cmd, "tags"),
						MinTraceCount: intPtrFlag(cmd, "min-trace-count"),
					},
				}
				res, err := app.client.CreateSessionRun(cmd.Context(), body)
				if err != nil {
					return err
				}
				return app.writer.Render(res)
			}

			status, _ := cmd.Flags().GetString("status")
			if err := validateEnum("status", status, models.TraceStatusValues()); err != nil {
				return err
			}
			body := models.CreateEvalRunRequest{
				Name:         strPtrFlag(cmd, "name"),
				Metrics:      metrics,
				SamplingRate: floatPtrFlag(cmd, "sampling-rate"),
				Model:        strPtrFlag(cmd, "model"),
				Filters: &models.EvalRunFilters{
					DateFrom:  strPtrFlag(cmd, "date-from"),
					DateTo:    strPtrFlag(cmd, "date-to"),
					Status:    strPtrFlag(cmd, "status"),
					SessionID: strPtrFlag(cmd, "session-id"),
					UserID:    strPtrFlag(cmd, "user-id"),
					Tags:      sliceFlag(cmd, "tags"),
					Name:      strPtrFlag(cmd, "filter-name"),
				},
			}
			res, err := app.client.CreateTraceRun(cmd.Context(), body)
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
	cmd.Flags().String("name", "", "Human-readable label for the run")
	cmd.Flags().StringSlice("metrics", nil, "Metric names to run (required, comma-separated)")
	cmd.Flags().Float64("sampling-rate", 0, "Fraction of targets to evaluate (0-1)")
	cmd.Flags().String("model", "", "Override the LLM judge model")
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

func newEvalsRunsBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "batch",
		Short:       "Create an evaluation run for an explicit set of IDs",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			metrics, _ := cmd.Flags().GetStringSlice("metrics")
			if len(metrics) == 0 {
				return exitcode.New(exitcode.Validation, "at least one --metrics value is required")
			}
			if target == targetSession {
				ids, _ := cmd.Flags().GetStringSlice("session-ids")
				if len(ids) == 0 {
					return exitcode.New(exitcode.Validation, "--session-ids is required for --target session")
				}
				weights, err := signalWeights(cmd)
				if err != nil {
					return err
				}
				res, err := app.client.BatchSessionRun(cmd.Context(), models.CreateBatchSessionEvalRunRequest{
					SessionIDs:    ids,
					Metrics:       metrics,
					Name:          strPtrFlag(cmd, "name"),
					Model:         strPtrFlag(cmd, "model"),
					SignalWeights: weights,
				})
				if err != nil {
					return err
				}
				return app.writer.Render(res)
			}
			ids, _ := cmd.Flags().GetStringSlice("trace-ids")
			if len(ids) == 0 {
				return exitcode.New(exitcode.Validation, "--trace-ids is required for --target trace")
			}
			res, err := app.client.BatchTraceRun(cmd.Context(), models.CreateBatchEvalRunRequest{
				TraceIDs: ids,
				Metrics:  metrics,
				Name:     strPtrFlag(cmd, "name"),
				Model:    strPtrFlag(cmd, "model"),
			})
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
	cmd.Flags().StringSlice("trace-ids", nil, "Trace IDs to evaluate (--target trace)")
	cmd.Flags().StringSlice("session-ids", nil, "Session IDs to evaluate (--target session)")
	cmd.Flags().StringSlice("metrics", nil, "Metric names to run (required, comma-separated)")
	cmd.Flags().String("name", "", "Human-readable label for the run")
	cmd.Flags().String("model", "", "Override the LLM judge model")
	cmd.Flags().String("signal-weights", "", "Session only: JSON object overriding signal weights")
	return cmd
}

func newEvalsRunsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List evaluation runs",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			q, err := newQuery(cmd).
				pagination(evalMaxLimit).
				enum("status", "status", models.EvaluationStatusValues()).
				build()
			if err != nil {
				return err
			}
			var res *models.Paginated[models.EvalRunResponse]
			if target == targetSession {
				res, err = app.client.ListSessionRuns(cmd.Context(), q)
			} else {
				res, err = app.client.ListTraceRuns(cmd.Context(), q)
			}
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().String("status", "", "Filter by run status (PENDING, RUNNING, COMPLETED, FAILED)")
	return cmd
}

func newEvalsRunsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "get <run_id>",
		Short:       "Get an evaluation run",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			var res *models.EvalRunResponse
			if target == targetSession {
				res, err = app.client.GetSessionRun(cmd.Context(), args[0])
			} else {
				res, err = app.client.GetTraceRun(cmd.Context(), args[0])
			}
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
}

func newEvalsRunsScoresCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "scores <run_id>",
		Short:       "List the scores produced by an evaluation run",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			if target == targetSession {
				res, e := app.client.SessionRunScores(cmd.Context(), args[0])
				if e != nil {
					return e
				}
				return app.writer.Render(*res)
			}
			res, e := app.client.TraceRunScores(cmd.Context(), args[0])
			if e != nil {
				return e
			}
			return app.writer.Render(*res)
		},
	}
}

func sliceFlag(cmd *cobra.Command, flag string) []string {
	f := cmd.Flags().Lookup(flag)
	if f == nil || !f.Changed {
		return nil
	}
	v, _ := cmd.Flags().GetStringSlice(flag)
	return v
}
