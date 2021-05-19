package action

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "action",
		Aliases: []string{"act"},
		Short:   "This command consists of multiple subcommands to interact with target Actions",
	}

	root.AddCommand(
		NewCreate(),
		NewDelete(),
		NewRun(),
		NewGet(),
		NewStatus(),
		NewWatch(),
	)
	return root
}
