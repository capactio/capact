package policy

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	och := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"pol"},
		Short:   "This command consists of multiple subcommands to interact with Policy",
	}

	och.AddCommand(
		NewGet(),
		NewEdit(),
		NewApply(),
	)
	return och
}
