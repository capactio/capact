package policy

import "github.com/spf13/cobra"

func NewPolicy() *cobra.Command {
	och := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"pol"},
		Short:   "This command consists of multiple subcommands to interact with Policy",
	}

	och.AddCommand(
		NewGet(),
		NewUpdate(),
	)
	return och
}
