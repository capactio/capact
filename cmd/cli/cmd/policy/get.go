package policy

import (
	"os"

	"capact.io/capact/internal/cli/policy"
	"github.com/spf13/cobra"
)

func NewGet() *cobra.Command {
	var opts policy.GetOptions

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Displays the details of current Policy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return policy.Get(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Output, "output", "o", "yaml", "Output format. One of:\njson | yaml")
	return cmd
}
