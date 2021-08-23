package implementations

import (
	"github.com/spf13/cobra"
)

// NewImplementations returns a cobra.Command for Hub Implementation related operations.
func NewImplementations() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "implementation",
		Aliases: []string{"impl", "implementations"},
		Short:   "This command consists of multiple subcommands to interact with Implementations stored on the Hub server",
	}

	cmd.AddCommand(
		NewGet(),
	)
	return cmd
}
