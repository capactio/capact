package hub

import (
	"capact.io/capact/cmd/cli/cmd/hub/implementations"
	"capact.io/capact/cmd/cli/cmd/hub/interfaces"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
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
