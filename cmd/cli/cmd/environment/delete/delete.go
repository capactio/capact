package delete

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "delete",
		Short: "This command consists of multiple subcommands to delete created Capact environment",
	}

	root.AddCommand(
		NewKind(),
	)
	return root
}
