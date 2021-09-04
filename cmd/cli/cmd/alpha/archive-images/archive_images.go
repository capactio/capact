package archiveimages

import (
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for archiving images operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive-images",
		Short: "Creates Docker images archive file",
		Long:  "Subcommand for various manifest generation operations",
	}

	cmd.AddCommand(NewFromHelmCharts())

	return cmd
}
