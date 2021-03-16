package implementations

import (
	"github.com/spf13/cobra"
)

func NewImplementations() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "implementations",
		Aliases: []string{"impl"},
		Short:   "This command consists of multiple subcommands to interact with Implementations stored on the Hub server",
	}

	cmd.AddCommand(
		NewSearch(),
	)
	return cmd
}
