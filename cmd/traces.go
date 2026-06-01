package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

const traceMaxLimit = 200

func newTracesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Inspect traces and their spans",
	}
	cmd.AddCommand(newTracesListCmd(), newTracesGetCmd(), newTracesSpansCmd())
	return cmd
}

func authAnnotation() map[string]string {
	return map[string]string{annotationRequiresAuth: "true"}
}

func newTracesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List traces with filtering and pagination",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			q, err := newQuery(cmd).
				pagination(traceMaxLimit).
				str("session_id", "session-id").
				enum("status", "status", models.TraceStatusValues()).
				str("user_id", "user-id").
				str("name", "name").
				strs("tags", "tags").
				date("started_after", "started-after").
				date("started_before", "started-before").
				enum("sort_by", "sort-by", models.TraceSortByValues()).
				enum("sort_order", "sort-order", models.SortOrderValues()).
				build()
			if err != nil {
				return err
			}
			res, err := app.client.ListTraces(cmd.Context(), q)
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().String("session-id", "", "Filter by session ID")
	cmd.Flags().String("status", "", "Filter by status (PENDING, RUNNING, COMPLETED, ERROR)")
	cmd.Flags().String("user-id", "", "Filter by user ID")
	cmd.Flags().String("name", "", "Filter by trace name (partial match)")
	cmd.Flags().StringSlice("tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().String("started-after", "", "Only traces started after this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("started-before", "", "Only traces started before this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("sort-by", "", "Sort field (started_at, ended_at, name, latency, status)")
	cmd.Flags().String("sort-order", "", "Sort order: asc or desc")
	return cmd
}

func newTracesGetCmd() *cobra.Command {
	var spansOnly bool
	var kind, status string
	cmd := &cobra.Command{
		Use:         "get <trace_id>",
		Short:       "Get a single trace, including its spans",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			if err := validateEnum("kind", kind, models.SpanKindValues()); err != nil {
				return err
			}
			if err := validateEnum("status", status, models.SpanStatusValues()); err != nil {
				return err
			}
			trace, err := app.client.GetTrace(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if spansOnly {
				return app.writer.Render(filterSpans(trace.Spans, kind, status))
			}
			return app.writer.Render(trace)
		},
	}
	cmd.Flags().BoolVar(&spansOnly, "spans-only", false, "Output only the spans array")
	cmd.Flags().StringVar(&kind, "kind", "", "When --spans-only, filter spans by kind")
	cmd.Flags().StringVar(&status, "status", "", "When --spans-only, filter spans by status")
	return cmd
}

func newTracesSpansCmd() *cobra.Command {
	var kind, status string
	cmd := &cobra.Command{
		Use:         "spans <trace_id>",
		Short:       "List the spans of a trace (filterable by kind/status)",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			if err := validateEnum("kind", kind, models.SpanKindValues()); err != nil {
				return err
			}
			if err := validateEnum("status", status, models.SpanStatusValues()); err != nil {
				return err
			}
			// Spans are inline in the trace response; filter client-side.
			trace, err := app.client.GetTrace(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return app.writer.Render(filterSpans(trace.Spans, kind, status))
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Filter by span kind (LLM, TOOL, AGENT, CHAIN, RETRIEVER, EMBEDDING, OTHER)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by span status (OK, ERROR, UNSET)")
	return cmd
}

func filterSpans(spans []models.SpanResponse, kind, status string) []models.SpanResponse {
	if kind == "" && status == "" {
		if spans == nil {
			return []models.SpanResponse{}
		}
		return spans
	}
	out := make([]models.SpanResponse, 0, len(spans))
	for _, s := range spans {
		if kind != "" && string(s.Kind) != kind {
			continue
		}
		if status != "" && string(s.Status) != status {
			continue
		}
		out = append(out, s)
	}
	return out
}
