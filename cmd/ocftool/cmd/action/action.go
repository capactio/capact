package action

import (
	"github.com/spf13/cobra"
)

func NewAction() *cobra.Command {
	och := &cobra.Command{
		Use:     "action",
		Aliases: []string{"act"},
		Short:   "This command consists of multiple subcommands to interact with target Actions",
	}

	och.AddCommand(
		NewCreate(),
		NewRun(),
		NewGet(),
		NewSearch(),
		NewStatus(),
		NewWatch(),
	)
	return och
}
