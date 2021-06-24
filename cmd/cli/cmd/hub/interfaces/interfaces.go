package interfaces

import (
	"github.com/spf13/cobra"
)

// NewInterfaces returns a cobra.Command for the "interfaces" command.
func NewInterfaces() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interfaces",
		Short: "This command consists of multiple subcommands to interact with Interfaces stored on the Hub server",
	}

	cmd.AddCommand(
		NewGet(),
		NewBrowse(),
	)
	return cmd
}
