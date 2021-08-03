package delete

import (
	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/environment/delete"
	"github.com/spf13/cobra"
)

// NewKind returns a cobra.Command for deleting kind environment.
func NewKind() *cobra.Command {
	var opts delete.KindOptions

	cmd := &cobra.Command{
		Use:   "kind",
		Short: "Delete local kind cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return delete.Kind(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", create.KindDefaultClusterName, "cluster name, overrides config")

	return cmd
}
