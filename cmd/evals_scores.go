package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

func newEvalsScoresCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scores",
		Short: "List, fetch and submit evaluation scores",
	}
	cmd.AddCommand(newEvalsScoresListCmd(), newEvalsScoresGetCmd(), newEvalsScoresSubmitCmd())
	return cmd
}

func newEvalsScoresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List scores with filtering",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			qb := newQuery(cmd).
				pagination(evalMaxLimit).
				str("name", "name").
				enum("source", "source", models.ScoreSourceValues()).
				enum("status", "status", models.ScoreStatusValues()).
				str("eval_run_id", "eval-run-id").
				date("date_from", "date-from").
				date("date_to", "date-to")

			if target == targetSession {
				qb = qb.str("session_id", "session-id")
				q, err := qb.build()
				if err != nil {
					return err
				}
				res, err := app.client.ListSessionScores(cmd.Context(), q)
				if err != nil {
					return err
				}
				return app.writer.Render(models.AsList(res))
			}

			qb = qb.str("trace_id", "trace-id").
				enum("data_type", "data-type", models.ScoreDataTypeValues()).
				str("environment", "environment")
			q, err := qb.build()
			if err != nil {
				return err
			}
			res, err := app.client.ListTraceScores(cmd.Context(), q)
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().String("trace-id", "", "Filter (trace): trace ID")
	cmd.Flags().String("session-id", "", "Filter (session): session ID")
	cmd.Flags().String("name", "", "Filter by metric name")
	cmd.Flags().String("source", "", "Filter by source (AUTOMATED, ANNOTATION, PROGRAMMATIC)")
	cmd.Flags().String("status", "", "Filter by status (SUCCESS, FAILED, PENDING)")
	cmd.Flags().String("data-type", "", "Filter (trace): data type (NUMERIC, BOOLEAN, CATEGORICAL)")
	cmd.Flags().String("eval-run-id", "", "Filter by evaluation run ID")
	cmd.Flags().String("environment", "", "Filter (trace): environment")
	cmd.Flags().String("date-from", "", "Filter: scores on/after this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("date-to", "", "Filter: scores on/before this date (RFC3339 or YYYY-MM-DD)")
	return cmd
}

func newEvalsScoresGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "get <trace_id|session_id>",
		Short:       "Get all scores for a specific trace or session",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			if target == targetSession {
				res, e := app.client.GetSessionScores(cmd.Context(), args[0])
				if e != nil {
					return e
				}
				return app.writer.Render(*res)
			}
			res, e := app.client.GetTraceScores(cmd.Context(), args[0])
			if e != nil {
				return e
			}
			return app.writer.Render(*res)
		},
	}
}

func newEvalsScoresSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "submit",
		Short:       "Submit a score for a trace (trace target only)",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			if target == targetSession {
				return exitcode.New(exitcode.Validation, "session score submission is not supported by the API; use --target trace")
			}

			traceID, _ := cmd.Flags().GetString("trace-id")
			name, _ := cmd.Flags().GetString("name")
			value, _ := cmd.Flags().GetString("value")
			if traceID == "" || name == "" || value == "" {
				return exitcode.New(exitcode.Validation, "--trace-id, --name and --value are all required")
			}

			dataType, _ := cmd.Flags().GetString("data-type")
			if err := validateEnum("data-type", dataType, models.ScoreDataTypeValues()); err != nil {
				return err
			}
			source, _ := cmd.Flags().GetString("source")
			if err := validateEnum("source", source, models.ScoreSourceValues()); err != nil {
				return err
			}

			var metadata models.JSON
			if raw, _ := cmd.Flags().GetString("metadata"); raw != "" {
				if !json.Valid([]byte(raw)) {
					return exitcode.New(exitcode.Validation, "invalid --metadata: must be valid JSON")
				}
				metadata = models.JSON(raw)
			}

			res, err := app.client.SubmitTraceScore(cmd.Context(), models.CreateTraceScoreRequest{
				TraceID:  traceID,
				Name:     name,
				Value:    value,
				DataType: dataType,
				Source:   source,
				Reason:   strPtrFlag(cmd, "reason"),
				Metadata: metadata,
			})
			if err != nil {
				return err
			}
			return app.writer.Render(res)
		},
	}
	cmd.Flags().String("trace-id", "", "Trace ID to score (required)")
	cmd.Flags().String("name", "", "Score name, e.g. accuracy (required)")
	cmd.Flags().String("value", "", "Score value as a string, e.g. 0.85 / true / PASS (required)")
	cmd.Flags().String("data-type", "NUMERIC", "Data type (NUMERIC, BOOLEAN, CATEGORICAL)")
	cmd.Flags().String("source", "PROGRAMMATIC", "Score source (AUTOMATED, ANNOTATION, PROGRAMMATIC)")
	cmd.Flags().String("reason", "", "Optional explanation for the score")
	cmd.Flags().String("metadata", "", "Optional JSON metadata object")
	return cmd
}
