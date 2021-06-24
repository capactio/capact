package implementations

import (
	"github.com/spf13/cobra"
)

// NewImplementations return a cobra.Command for the "implementations" subcommand.
func NewImplementations() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "implementations",
		Aliases: []string{"impl"},
		Short:   "This command consists of multiple subcommands to interact with Implementations stored on the Hub server",
	}

	cmd.AddCommand(
		NewGet(),
	)
	return cmd
}
