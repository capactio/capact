package policy

import (
	"os"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/policy"
	"capact.io/capact/internal/cli/printer"

	"github.com/spf13/cobra"
)

// NewGet return a cobra.Command for getting the Capact Global policy on a Capact environment.
func NewGet() *cobra.Command {
	resourcePrinter := printer.NewForResource(
		os.Stdout,
		printer.WithJSON(),
		printer.WithJSONPath(),
		printer.WithYAML(),
		printer.WithDefaultOutputFormat(printer.YAMLFormat),
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Displays the details of current Policy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return policy.Get(cmd.Context(), resourcePrinter)
		},
	}

	flags := cmd.Flags()
	resourcePrinter.RegisterFlags(flags)
	client.RegisterFlags(flags)

	return cmd
}
