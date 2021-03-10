package action

import (
	"github.com/spf13/cobra"
)

func NewAction() *cobra.Command {
	och := &cobra.Command{
		Use:   "action",
		Short: "This command consists of multiple subcommands to interact with Action.",
	}

	och.AddCommand(NewCreate())
	return och
}
