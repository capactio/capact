package archiveimages

import (
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for archiving images operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive-images",
		Short: "Export Capact Docker images to a tar archive",
		Long:  "Subcommand for various manifest generation operations",
	}

	cmd.AddCommand(NewFromHelmCharts())

	return cmd
}
