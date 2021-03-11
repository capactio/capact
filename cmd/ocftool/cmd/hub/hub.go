package hub

import (
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/hub/implementations"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/hub/interfaces"

	"github.com/spf13/cobra"
)

func NewHub() *cobra.Command {
	och := &cobra.Command{
		Use:   "hub",
		Short: "This command consists of multiple subcommands to interact with Hub server.",
	}

	och.AddCommand(
		interfaces.NewInterfaces(),
		implementations.NewImplementations(),
	)

	return och
}
