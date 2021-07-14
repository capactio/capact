package delete

import (
	"capact.io/capact/internal/cli/environment/delete"
	"github.com/spf13/cobra"
)

// NewKind returns a cobra.Command for deleting Kind environment.
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
	cmd.Flags().StringVar(&opts.Name, "name", "kind-dev-capact", "cluster name, overrides config")

	return cmd
}
