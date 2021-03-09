package och

import (
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/och/list"

	"github.com/spf13/cobra"
)

func NewOCH() *cobra.Command {
	och := &cobra.Command{
		Use:   "och",
		Short: "This command consists of multiple subcommands to interact with OCH server.",
	}

	och.AddCommand(list.NewCmdList())
	return och
}
