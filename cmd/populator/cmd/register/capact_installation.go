package register

import (
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"

	"projectvoltron.dev/voltron/internal/installation"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"
)

// TODO: support configuration both via flags and environment variables
func NewCapactInstallation(cliName string) *cobra.Command {
	return &cobra.Command{
		Use:   "capact-installation",
		Short: "Produces and uploads TypeInstances which describe Capact installation",
		Example: heredoc.WithCLIName(`
			<cli> capact-installation
		`, cliName),
		Args: cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			capactRegister, err := installation.NewCapactRegister()
			if err != nil {
				return err
			}
			return capactRegister.RegisterTypeInstances(cmd.Context())
		},
	}
}
