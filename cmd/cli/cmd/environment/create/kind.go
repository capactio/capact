package create

import (
	"time"

	"capact.io/capact/internal/cli/environment/create"
	"github.com/spf13/cobra"
)

// NewKind returns a cobra.Command for creating kind environment.
func NewKind() *cobra.Command {
	var opts create.KindOptions

	cmd := &cobra.Command{
		Use:   "kind",
		Short: "Provision local kind cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return create.Kind(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", create.DefaultClusterName, "cluster name, overrides config")
	cmd.Flags().StringVar(&opts.Config, "cluster-config", "", "path to a kind config file")
	cmd.Flags().StringVar(&opts.ImageName, "image", create.KindDefaultNodeImage, "node docker image to use for booting the cluster")
	cmd.Flags().BoolVar(&opts.Retain, "retain", false, "retain nodes for debugging when cluster creation fails")
	cmd.Flags().DurationVar(&opts.Wait, "wait", time.Duration(0), "wait for control plane node to be ready")
	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", "", "sets kubeconfig path instead of $KUBECONFIG or $HOME/.kube/config")

	return cmd
}
