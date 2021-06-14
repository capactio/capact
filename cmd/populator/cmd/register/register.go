package register

import (
	"github.com/spf13/cobra"
)

func NewRegister(cliName string) *cobra.Command {
	hub := &cobra.Command{
		Use:   "register",
		Short: "This command consists of multiple subcommands which allows you to register Capact resources",
	}

	hub.AddCommand(
		NewCapactInstallation(cliName),
		NewOCFManifests(cliName),
	)
	return hub
}
