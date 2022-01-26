package alpha

import (
	archiveimages "capact.io/capact/cmd/cli/cmd/alpha/archive-images"
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for operations, which are in alpha version.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Alpha features",
		Long:  "Subcommand for alpha features in the CLI",
	}

	cmd.AddCommand(
		archiveimages.NewCmd(),
	)

	return cmd
}
