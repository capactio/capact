package cmd

import (
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"

	"projectvoltron.dev/voltron/internal/ocftool/heredoc"
	"projectvoltron.dev/voltron/internal/publisher"
)

// TODO: support configuration both via flags and environment variables
func NewInstallTypeInstances() *cobra.Command {
	return &cobra.Command{
		Use:   "voltron-install-type-instances",
		Short: "Produces and uploads TypeInstances which describe Voltron installation",
		Example: heredoc.WithCLIName(`
			<cli> voltron-install-type-instances
		`, CLIName),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pub, err := publisher.NewTypeInstances()
			if err != nil {
				return err
			}
			return pub.PublishVoltronInstallTypeInstances(cmd.Context())
		},
	}
}
