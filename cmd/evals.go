package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/exitcode"
)

const (
	targetTrace   = "trace"
	targetSession = "session"
	evalMaxLimit  = 200
)

func newEvalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evals",
		Short: "Manage evaluations: metrics, runs, scores and monitors",
		Long:  "Evaluation commands operate on either traces or sessions. Use --target to choose\n(default: trace). Some operations are trace-only and will error for --target session.",
	}
	// --target is shared by every evals subcommand.
	cmd.PersistentFlags().String("target", targetTrace, "Evaluation target: trace or session")
	cmd.AddCommand(newEvalsMetricsCmd(), newEvalsRunsCmd(), newEvalsScoresCmd(), newEvalsMonitorsCmd())
	return cmd
}

// evalTarget reads and validates the inherited --target flag.
func evalTarget(cmd *cobra.Command) (string, error) {
	t, _ := cmd.Flags().GetString("target")
	if t != targetTrace && t != targetSession {
		return "", exitcode.New(exitcode.Validation, "invalid --target %q: must be \"trace\" or \"session\"", t)
	}
	return t, nil
}
