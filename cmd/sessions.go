package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

const (
	sessionListMaxLimit = 200
	sessionGetMaxLimit  = 1000
)

func newSessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Inspect sessions and their traces",
	}
	cmd.AddCommand(newSessionsListCmd(), newSessionsGetCmd())
	return cmd
}

func newSessionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List sessions with filtering and pagination",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			q, err := newQuery(cmd).
				pagination(sessionListMaxLimit).
				str("user_id", "user-id").
				boolean("has_error", "has-error").
				date("started_after", "started-after").
				date("started_before", "started-before").
				strs("tags", "tags").
				str("query", "query").
				enum("sort_by", "sort-by", models.SessionSortByValues()).
				enum("sort_order", "sort-order", models.SortOrderValues()).
				build()
			if err != nil {
				return err
			}
			res, err := app.client.ListSessions(cmd.Context(), q)
			if err != nil {
				return err
			}
			return app.writer.Render(models.AsList(res))
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().String("user-id", "", "Filter by user ID")
	cmd.Flags().Bool("has-error", false, "Filter by presence of an error")
	cmd.Flags().String("started-after", "", "Only sessions started after this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("started-before", "", "Only sessions started before this date (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringSlice("tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().String("query", "", "Substring filter on session ID")
	cmd.Flags().String("sort-by", "", "Sort field (recent, trace_count, latency, cost)")
	cmd.Flags().String("sort-order", "", "Sort order: asc or desc")
	return cmd
}

func newSessionsGetCmd() *cobra.Command {
	var includeTraces bool
	cmd := &cobra.Command{
		Use:         "get <session_id>",
		Short:       "Get a session, including its traces",
		Args:        cobra.ExactArgs(1),
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := appFrom(cmd)
			q, err := newQuery(cmd).pagination(sessionGetMaxLimit).build()
			if err != nil {
				return err
			}
			session, err := app.client.GetSession(cmd.Context(), args[0], q)
			if err != nil {
				return err
			}
			if !includeTraces {
				// The API always returns traces; this trims them client-side to
				// reduce output size.
				session.Traces = nil
			}
			return app.writer.Render(session)
		},
	}
	addPaginationFlags(cmd)
	cmd.Flags().BoolVar(&includeTraces, "include-traces", true, "Include the session's traces in the output")
	return cmd
}
