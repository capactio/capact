package action

import (
	"github.com/spf13/cobra"
)

func NewAction() *cobra.Command {
	och := &cobra.Command{
		Use:     "action",
		Aliases: []string{"act"},
		Short:   "This command consists of multiple subcommands to interact with Action.",
	}

	och.AddCommand(
		NewCreate(),
		NewRun(),
		NewGet(),
		NewSearch(),
	)
	return och
}
