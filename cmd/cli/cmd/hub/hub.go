package hub

import (
	"capact.io/capact/cmd/cli/cmd/hub/implementations"
	"capact.io/capact/cmd/cli/cmd/hub/interfaces"

	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for interacting with the Hub.
func NewCmd() *cobra.Command {
	hub := &cobra.Command{
		Use:   "hub",
		Short: "This command consists of multiple subcommands to interact with Hub server.",
	}

	hub.AddCommand(
		interfaces.NewInterfaces(),
		implementations.NewImplementations(),
	)

	return hub
}
