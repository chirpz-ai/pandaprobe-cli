package cmd

import "github.com/spf13/cobra"

func newEvalsMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "metrics",
		Short:       "List evaluation metrics available for the target",
		Args:        cobra.NoArgs,
		Annotations: authAnnotation(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			app := appFrom(cmd)
			target, err := evalTarget(cmd)
			if err != nil {
				return err
			}
			var metrics any
			if target == targetSession {
				m, e := app.client.SessionMetrics(cmd.Context())
				if e != nil {
					return e
				}
				metrics = *m
			} else {
				m, e := app.client.TraceMetrics(cmd.Context())
				if e != nil {
					return e
				}
				metrics = *m
			}
			return app.writer.Render(metrics)
		},
	}
}
