package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chirpz-ai/pandaprobe-cli/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return appFrom(cmd).writer.Render(version.Get())
		},
	}
}
