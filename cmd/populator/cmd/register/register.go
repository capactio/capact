package register

import (
	"github.com/spf13/cobra"
)

// NewRegister returns a cobra.Command subcommand for registering Capact resources.
func NewRegister(cliName string) *cobra.Command {
	hub := &cobra.Command{
		Use:   "register",
		Short: "This command consists of multiple subcommands which allows you to register Capact resources",
	}

	hub.AddCommand(
		NewCapactInstallation(cliName),
		NewTestStorageBackend(cliName),
		NewOCFManifests(cliName),
	)
	return hub
}
