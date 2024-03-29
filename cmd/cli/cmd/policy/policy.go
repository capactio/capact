package policy

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for policy related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"pol"},
		Short:   "This command consists of multiple subcommands to interact with Policy",
	}

	root.AddCommand(
		NewGet(),
		NewEdit(),
		NewApply(),
	)
	return root
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
