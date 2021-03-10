package interfaces

import (
	"github.com/spf13/cobra"
)

func NewInterfaces() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interfaces",
		Short: "This command consists of multiple subcommands to interact with OCH server.",
	}

	cmd.AddCommand(
		NewSearch(),
		NewBrowse(),
	)
	return cmd
}
