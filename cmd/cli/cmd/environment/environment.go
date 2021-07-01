package environment

import (
	"capact.io/capact/cmd/cli/cmd/environment/create"
	deletecluster "capact.io/capact/cmd/cli/cmd/environment/delete"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	hub := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   "This command consists of multiple subcommands to interact with a Kubernetes cluster",
	}

	hub.AddCommand(
		create.NewCmd(),
		deletecluster.NewCmd(),
	)

	return hub
}
